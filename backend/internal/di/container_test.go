package di

import (
	"config-service/backend/internal/infrastructure/database"
	"config-service/backend/internal/model"
	"config-service/backend/internal/repository"
	"config-service/backend/pkg/metrics"
	"database/sql"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type diStubRepository struct{}

func (diStubRepository) Create(*model.Config) error {
	return nil
}

func (diStubRepository) Get(environment, key string) (*model.Config, error) {
	return &model.Config{Environment: environment, Key: key, Value: "value"}, nil
}

func (diStubRepository) GetAll(environment string) ([]*model.Config, error) {
	return []*model.Config{{Environment: environment, Key: "key", Value: "value"}}, nil
}

func (diStubRepository) Update(*model.Config) error {
	return nil
}

func (diStubRepository) Delete(string, string) error {
	return nil
}

func (diStubRepository) Exists(string, string) (bool, error) {
	return false, nil
}

type diStubConnection struct {
	db *sql.DB
}

func (c diStubConnection) GetDB() *sql.DB {
	return c.db
}

func (c diStubConnection) Close() error {
	return nil
}

func diTestMetrics() *metrics.Metrics {
	return &metrics.Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "di_test_http_requests_total"},
			[]string{"path", "method", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "di_test_http_request_duration_seconds"},
			[]string{"path", "method"},
		),
		DBQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "di_test_db_queries_total"},
			[]string{"operation"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "di_test_db_query_duration_seconds"},
			[]string{"operation"},
		),
	}
}

func TestProviderHelpers(t *testing.T) {
	var repo repository.ConfigRepository = diStubRepository{}
	svc := provideConfigService(repo)
	if svc == nil {
		t.Fatal("provideConfigService() returned nil")
	}

	h := provideConfigHandler(svc)
	if h == nil {
		t.Fatal("provideConfigHandler() returned nil")
	}
}

func TestProvideConfigRepository(t *testing.T) {
	conn := diStubConnection{db: nil}

	repo, err := provideConfigRepository(conn, diTestMetrics())
	if err != nil {
		t.Fatalf("provideConfigRepository() error = %v", err)
	}
	if repo == nil {
		t.Fatal("provideConfigRepository() returned nil")
	}

	var _ database.Connection = conn
}
