package database

import (
	"config-service/backend/internal/model"
	"config-service/backend/internal/repository"
	"database/sql"
	"embed"
	"errors"
	"strings"
	"time"
)

//go:embed queries/*.sql
var queriesFS embed.FS

type postgresRepository struct {
	db      *sql.DB
	queries map[string]string
}

func NewPostgresRepository(db *sql.DB) (repository.ConfigRepository, error) {
	queries, err := loadQueries()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{
		db:      db,
		queries: queries,
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
	query := r.queries["create_config"]
	if query == "" {
		return errors.New("create_config query not found")
	}
	_, err := r.db.Exec(query, config.Environment, config.Key, config.Value, config.UpdatedAt)
	return err
}

func (r *postgresRepository) Get(environment, key string) (*model.Config, error) {
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
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("config not found")
		}
		return nil, err
	}

	config.UpdatedAt = updatedAt
	return &config, nil
}

func (r *postgresRepository) GetAll(environment string) ([]*model.Config, error) {
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

	return configs, nil
}

func (r *postgresRepository) Update(config *model.Config) error {
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
		return errors.New("config not found")
	}

	return nil
}

func (r *postgresRepository) Delete(environment, key string) error {
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
		return errors.New("config not found")
	}

	return nil
}

func (r *postgresRepository) Exists(environment, key string) (bool, error) {
	query := r.queries["exists_config"]
	if query == "" {
		return false, errors.New("exists_config query not found")
	}
	var exists bool
	err := r.db.QueryRow(query, environment, key).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
