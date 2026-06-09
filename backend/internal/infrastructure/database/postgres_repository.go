package database

import (
	"config-service/backend/internal/model"
	"config-service/backend/internal/repository"
	"config-service/backend/pkg/metrics"
	"database/sql"
	"embed"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

//go:embed queries/*.sql
var queriesFS embed.FS

type postgresRepository struct {
	db      *sql.DB
	queries map[string]string
	metrics *metrics.Metrics
}

func NewPostgresRepository(db *sql.DB, m *metrics.Metrics) (repository.ConfigRepository, error) {
	queries, err := loadQueries()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{
		db:      db,
		queries: queries,
		metrics: m,
	}, nil
}

func loadQueries() (map[string]string, error) {
	queries := make(map[string]string)
	entries, err := queriesFS.ReadDir("queries")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		data, err := queriesFS.ReadFile("queries/" + name)
		if err != nil {
			return nil, err
		}
		queryName := name[:len(name)-4]
		queries[queryName] = strings.TrimSpace(string(data))
	}

	return queries, nil
}

func (r *postgresRepository) Create(config *model.Config) error {
	start := time.Now()
	query := r.queries["create_config"]
	if query == "" {
		return errors.New("create_config query not found")
	}
	_, err := r.db.Exec(query, config.Environment, config.Key, config.Value, config.UpdatedAt)
	duration := time.Since(start).Seconds()
	r.metrics.DBQueriesTotal.WithLabelValues("create").Inc()
	r.metrics.DBQueryDuration.WithLabelValues("create").Observe(duration)
	if isUniqueViolation(err) {
		return repository.ErrConfigAlreadyExists
	}
	return err
}

func (r *postgresRepository) Get(environment, key string) (*model.Config, error) {
	start := time.Now()
	query := r.queries["get_config"]
	if query == "" {
		return nil, errors.New("get_config query not found")
	}
	var config model.Config
	var updatedAt time.Time

	err := r.db.QueryRow(query, environment, key).Scan(
		&config.Environment,
		&config.Key,
		&config.Value,
		&updatedAt,
	)
	duration := time.Since(start).Seconds()
	r.metrics.DBQueriesTotal.WithLabelValues("get").Inc()
	r.metrics.DBQueryDuration.WithLabelValues("get").Observe(duration)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrConfigNotFound
		}
		return nil, err
	}

	config.UpdatedAt = updatedAt
	return &config, nil
}

func (r *postgresRepository) GetAll(environment string) ([]*model.Config, error) {
	start := time.Now()
	query := r.queries["get_all_configs"]
	if query == "" {
		return nil, errors.New("get_all_configs query not found")
	}
	rows, err := r.db.Query(query, environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*model.Config
	for rows.Next() {
		var config model.Config
		var updatedAt time.Time
		if err := rows.Scan(
			&config.Environment,
			&config.Key,
			&config.Value,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		config.UpdatedAt = updatedAt
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	duration := time.Since(start).Seconds()
	r.metrics.DBQueriesTotal.WithLabelValues("get_all").Inc()
	r.metrics.DBQueryDuration.WithLabelValues("get_all").Observe(duration)
	return configs, nil
}

func (r *postgresRepository) Update(config *model.Config) error {
	start := time.Now()
	query := r.queries["update_config"]
	if query == "" {
		return errors.New("update_config query not found")
	}
	result, err := r.db.Exec(query, config.Environment, config.Key, config.Value, config.UpdatedAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrConfigNotFound
	}

	duration := time.Since(start).Seconds()
	r.metrics.DBQueriesTotal.WithLabelValues("update").Inc()
	r.metrics.DBQueryDuration.WithLabelValues("update").Observe(duration)
	return nil
}

func (r *postgresRepository) Delete(environment, key string) error {
	start := time.Now()
	query := r.queries["delete_config"]
	if query == "" {
		return errors.New("delete_config query not found")
	}
	result, err := r.db.Exec(query, environment, key)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrConfigNotFound
	}

	duration := time.Since(start).Seconds()
	r.metrics.DBQueriesTotal.WithLabelValues("delete").Inc()
	r.metrics.DBQueryDuration.WithLabelValues("delete").Observe(duration)
	return nil
}

func (r *postgresRepository) Exists(environment, key string) (bool, error) {
	start := time.Now()
	query := r.queries["exists_config"]
	if query == "" {
		return false, errors.New("exists_config query not found")
	}
	var exists bool
	err := r.db.QueryRow(query, environment, key).Scan(&exists)
	duration := time.Since(start).Seconds()
	r.metrics.DBQueriesTotal.WithLabelValues("exists").Inc()
	r.metrics.DBQueryDuration.WithLabelValues("exists").Observe(duration)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
