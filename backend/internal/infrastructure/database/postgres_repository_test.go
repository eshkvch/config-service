package database

import (
	"config-service/backend/internal/model"
	"config-service/backend/pkg/metrics"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var fakeDriverID atomic.Int64

type fakeDriver struct {
	state *fakeDBState
}

type fakeDBState struct {
	execResult driver.Result
	execErr    error
	queryRows  *fakeRows
	queryErr   error
}

type fakeConn struct {
	state *fakeDBState
}

type fakeRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	err     error
}

type fakeResult struct {
	rowsAffected int64
	rowsErr      error
}

func (d fakeDriver) Open(string) (driver.Conn, error) {
	return &fakeConn{state: d.state}, nil
}

func (c *fakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}

func (c *fakeConn) Close() error {
	return nil
}

func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported")
}

func (c *fakeConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	if c.state.execErr != nil {
		return nil, c.state.execErr
	}
	if c.state.execResult != nil {
		return c.state.execResult, nil
	}
	return fakeResult{rowsAffected: 1}, nil
}

func (c *fakeConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.state.queryErr != nil {
		return nil, c.state.queryErr
	}
	if c.state.queryRows != nil {
		return c.state.queryRows, nil
	}
	return &fakeRows{columns: []string{"exists"}, values: [][]driver.Value{{true}}}, nil
}

func (r fakeRows) Columns() []string {
	return r.columns
}

func (r *fakeRows) Close() error {
	return nil
}

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		if r.err != nil {
			return r.err
		}
		return io.EOF
	}

	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func (r fakeResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r fakeResult) RowsAffected() (int64, error) {
	return r.rowsAffected, r.rowsErr
}

func newFakeDB(t *testing.T, state *fakeDBState) *sql.DB {
	t.Helper()

	name := fmt.Sprintf("fake-postgres-%d", fakeDriverID.Add(1))
	sql.Register(name, fakeDriver{state: state})

	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func newRepositoryForTest(t *testing.T, state *fakeDBState) *postgresRepository {
	t.Helper()

	repo, err := NewPostgresRepository(newFakeDB(t, state), newRepositoryMetrics())
	if err != nil {
		t.Fatalf("NewPostgresRepository() error = %v", err)
	}
	return repo.(*postgresRepository)
}

func newRepositoryMetrics() *metrics.Metrics {
	return &metrics.Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "test_http_requests_total"},
			[]string{"path", "method", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "test_http_request_duration_seconds"},
			[]string{"path", "method"},
		),
		DBQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "test_db_queries_total"},
			[]string{"operation"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "test_db_query_duration_seconds"},
			[]string{"operation"},
		),
	}
}

func TestLoadQueries(t *testing.T) {
	queries, err := loadQueries()
	if err != nil {
		t.Fatalf("loadQueries() error = %v", err)
	}

	for _, name := range []string{
		"create_config",
		"delete_config",
		"exists_config",
		"get_all_configs",
		"get_config",
		"update_config",
	} {
		if strings.TrimSpace(queries[name]) == "" {
			t.Fatalf("query %q is missing", name)
		}
	}
}

func TestNewPostgresRepository(t *testing.T) {
	repo, err := NewPostgresRepository(newFakeDB(t, &fakeDBState{}), newRepositoryMetrics())
	if err != nil {
		t.Fatalf("NewPostgresRepository() error = %v", err)
	}
	if repo == nil {
		t.Fatal("repository is nil")
	}
}

func TestPostgresRepositoryMissingQueries(t *testing.T) {
	repo := &postgresRepository{
		db:      newFakeDB(t, &fakeDBState{}),
		queries: map[string]string{},
		metrics: newRepositoryMetrics(),
	}
	config := &model.Config{Environment: "prod", Key: "key", Value: "value", UpdatedAt: time.Now()}

	if err := repo.Create(config); err == nil || !strings.Contains(err.Error(), "create_config") {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := repo.Get("prod", "key"); err == nil || !strings.Contains(err.Error(), "get_config") {
		t.Fatalf("Get() error = %v", err)
	}
	if _, err := repo.GetAll("prod"); err == nil || !strings.Contains(err.Error(), "get_all_configs") {
		t.Fatalf("GetAll() error = %v", err)
	}
	if err := repo.Update(config); err == nil || !strings.Contains(err.Error(), "update_config") {
		t.Fatalf("Update() error = %v", err)
	}
	if err := repo.Delete("prod", "key"); err == nil || !strings.Contains(err.Error(), "delete_config") {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.Exists("prod", "key"); err == nil || !strings.Contains(err.Error(), "exists_config") {
		t.Fatalf("Exists() error = %v", err)
	}
}

func TestPostgresRepositoryCreate(t *testing.T) {
	config := &model.Config{Environment: "prod", Key: "key", Value: "value", UpdatedAt: time.Now()}

	if err := newRepositoryForTest(t, &fakeDBState{}).Create(config); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	wantErr := errors.New("exec failed")
	if err := newRepositoryForTest(t, &fakeDBState{execErr: wantErr}).Create(config); !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want %v", err, wantErr)
	}
}

func TestPostgresRepositoryGet(t *testing.T) {
	updatedAt := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	repo := newRepositoryForTest(t, &fakeDBState{
		queryRows: &fakeRows{
			columns: []string{"env", "key", "value", "updated_at"},
			values:  [][]driver.Value{{"prod", "key", "value", updatedAt}},
		},
	})

	config, err := repo.Get("prod", "key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if config.Environment != "prod" || config.Key != "key" || config.Value != "value" || !config.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("Get() = %#v", config)
	}

	_, err = newRepositoryForTest(t, &fakeDBState{
		queryRows: &fakeRows{columns: []string{"env", "key", "value", "updated_at"}},
	}).Get("prod", "missing")
	if err == nil || !strings.Contains(err.Error(), "config not found") {
		t.Fatalf("Get() no rows error = %v", err)
	}

	wantErr := errors.New("query failed")
	_, err = newRepositoryForTest(t, &fakeDBState{queryErr: wantErr}).Get("prod", "key")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Get() error = %v, want %v", err, wantErr)
	}
}

func TestPostgresRepositoryGetAll(t *testing.T) {
	updatedAt := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	repo := newRepositoryForTest(t, &fakeDBState{
		queryRows: &fakeRows{
			columns: []string{"env", "key", "value", "updated_at"},
			values: [][]driver.Value{
				{"prod", "a", "1", updatedAt},
				{"prod", "b", "2", updatedAt.Add(time.Minute)},
			},
		},
	})

	configs, err := repo.GetAll("prod")
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(configs) != 2 || configs[0].Key != "a" || configs[1].Key != "b" {
		t.Fatalf("GetAll() = %#v", configs)
	}

	wantErr := errors.New("query failed")
	_, err = newRepositoryForTest(t, &fakeDBState{queryErr: wantErr}).GetAll("prod")
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetAll() query error = %v, want %v", err, wantErr)
	}

	_, err = newRepositoryForTest(t, &fakeDBState{
		queryRows: &fakeRows{
			columns: []string{"env", "key", "value", "updated_at"},
			values:  [][]driver.Value{{"prod", "a", "1", "not-a-time"}},
		},
	}).GetAll("prod")
	if err == nil {
		t.Fatal("expected scan error")
	}

	_, err = newRepositoryForTest(t, &fakeDBState{
		queryRows: &fakeRows{
			columns: []string{"env", "key", "value", "updated_at"},
			err:     errors.New("rows failed"),
		},
	}).GetAll("prod")
	if err == nil || !strings.Contains(err.Error(), "rows failed") {
		t.Fatalf("GetAll() rows error = %v", err)
	}
}

func TestPostgresRepositoryUpdate(t *testing.T) {
	config := &model.Config{Environment: "prod", Key: "key", Value: "value", UpdatedAt: time.Now()}

	if err := newRepositoryForTest(t, &fakeDBState{}).Update(config); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := newRepositoryForTest(t, &fakeDBState{
		execResult: fakeResult{rowsAffected: 0},
	}).Update(config); err == nil || !strings.Contains(err.Error(), "config not found") {
		t.Fatalf("Update() no rows error = %v", err)
	}

	wantErr := errors.New("exec failed")
	if err := newRepositoryForTest(t, &fakeDBState{execErr: wantErr}).Update(config); !errors.Is(err, wantErr) {
		t.Fatalf("Update() exec error = %v, want %v", err, wantErr)
	}

	rowsErr := errors.New("rows affected failed")
	if err := newRepositoryForTest(t, &fakeDBState{
		execResult: fakeResult{rowsAffected: 1, rowsErr: rowsErr},
	}).Update(config); !errors.Is(err, rowsErr) {
		t.Fatalf("Update() rows error = %v, want %v", err, rowsErr)
	}
}

func TestPostgresRepositoryDelete(t *testing.T) {
	if err := newRepositoryForTest(t, &fakeDBState{}).Delete("prod", "key"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if err := newRepositoryForTest(t, &fakeDBState{
		execResult: fakeResult{rowsAffected: 0},
	}).Delete("prod", "key"); err == nil || !strings.Contains(err.Error(), "config not found") {
		t.Fatalf("Delete() no rows error = %v", err)
	}

	wantErr := errors.New("exec failed")
	if err := newRepositoryForTest(t, &fakeDBState{execErr: wantErr}).Delete("prod", "key"); !errors.Is(err, wantErr) {
		t.Fatalf("Delete() exec error = %v, want %v", err, wantErr)
	}

	rowsErr := errors.New("rows affected failed")
	if err := newRepositoryForTest(t, &fakeDBState{
		execResult: fakeResult{rowsAffected: 1, rowsErr: rowsErr},
	}).Delete("prod", "key"); !errors.Is(err, rowsErr) {
		t.Fatalf("Delete() rows error = %v, want %v", err, rowsErr)
	}
}

func TestPostgresRepositoryExists(t *testing.T) {
	exists, err := newRepositoryForTest(t, &fakeDBState{
		queryRows: &fakeRows{
			columns: []string{"exists"},
			values:  [][]driver.Value{{true}},
		},
	}).Exists("prod", "key")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Fatal("Exists() = false, want true")
	}

	wantErr := errors.New("query failed")
	_, err = newRepositoryForTest(t, &fakeDBState{queryErr: wantErr}).Exists("prod", "key")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Exists() error = %v, want %v", err, wantErr)
	}
}

func TestPostgresConnectionAccessors(t *testing.T) {
	db := newFakeDB(t, &fakeDBState{})
	conn := &postgresConnection{db: db}

	if conn.GetDB() != db {
		t.Fatal("GetDB() did not return wrapped DB")
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNewPostgresConnectionReturnsPingError(t *testing.T) {
	conn, err := NewPostgresConnection("invalid dsn")
	if err == nil {
		if conn != nil {
			_ = conn.Close()
		}
		t.Fatal("expected ping error")
	}
	if !strings.Contains(err.Error(), "failed to ping database") {
		t.Fatalf("unexpected error: %v", err)
	}
}
