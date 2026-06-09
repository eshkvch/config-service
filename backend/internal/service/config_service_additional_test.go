package service

import (
	"config-service/backend/internal/model"
	"config-service/backend/internal/repository"
	"errors"
	"testing"
)

type controllableRepository struct {
	createErr error
	getConfig *model.Config
	getErr    error
	getAll    []*model.Config
	getAllErr error
	updateErr error
	deleteErr error
	exists    bool
	existsErr error

	created    *model.Config
	updated    *model.Config
	deletedEnv string
	deletedKey string
	existsEnv  string
	existsKey  string
	getAllEnv  string
}

func (r *controllableRepository) Create(config *model.Config) error {
	r.created = config
	return r.createErr
}

func (r *controllableRepository) Get(environment, key string) (*model.Config, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.getConfig, nil
}

func (r *controllableRepository) GetAll(environment string) ([]*model.Config, error) {
	r.getAllEnv = environment
	return r.getAll, r.getAllErr
}

func (r *controllableRepository) Update(config *model.Config) error {
	r.updated = config
	return r.updateErr
}

func (r *controllableRepository) Delete(environment, key string) error {
	r.deletedEnv = environment
	r.deletedKey = key
	return r.deleteErr
}

func (r *controllableRepository) Exists(environment, key string) (bool, error) {
	r.existsEnv = environment
	r.existsKey = key
	return r.exists, r.existsErr
}

func TestConfigService_CreateConfigRepositoryErrors(t *testing.T) {
	tests := []struct {
		name    string
		repo    *controllableRepository
		wantErr error
	}{
		{
			name:    "exists error",
			repo:    &controllableRepository{existsErr: errors.New("db down")},
			wantErr: errors.New("db down"),
		},
		{
			name:    "create conflict",
			repo:    &controllableRepository{createErr: repository.ErrConfigAlreadyExists},
			wantErr: ErrConfigExists,
		},
		{
			name:    "create database error",
			repo:    &controllableRepository{createErr: errors.New("insert failed")},
			wantErr: errors.New("insert failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigService(tt.repo).CreateConfig("prod", "key", "value")
			if err == nil {
				t.Fatal("expected error")
			}
			if tt.wantErr == ErrConfigExists {
				if !errors.Is(err, ErrConfigExists) {
					t.Fatalf("CreateConfig() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err.Error() != tt.wantErr.Error() {
				t.Fatalf("CreateConfig() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigService_GetConfigReturnsDatabaseError(t *testing.T) {
	wantErr := errors.New("select failed")
	repo := &controllableRepository{getErr: wantErr}

	_, err := NewConfigService(repo).GetConfig("prod", "key")
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetConfig() error = %v, want %v", err, wantErr)
	}
}

func TestConfigService_GetAllConfigs(t *testing.T) {
	config, err := model.NewConfig("prod", "key", "value")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	repo := &controllableRepository{getAll: []*model.Config{config}}
	got, err := NewConfigService(repo).GetAllConfigs("prod")
	if err != nil {
		t.Fatalf("GetAllConfigs() error = %v", err)
	}
	if len(got) != 1 || got[0].Key != "key" {
		t.Fatalf("GetAllConfigs() = %#v", got)
	}
	if repo.getAllEnv != "prod" {
		t.Fatalf("GetAll called with env %q", repo.getAllEnv)
	}
}

func TestConfigService_GetAllConfigsReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("select failed")
	repo := &controllableRepository{getAllErr: wantErr}

	_, err := NewConfigService(repo).GetAllConfigs("prod")
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetAllConfigs() error = %v, want %v", err, wantErr)
	}
}

func TestConfigService_UpdateConfigErrors(t *testing.T) {
	config, err := model.NewConfig("prod", "key", "old")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	tests := []struct {
		name    string
		repo    *controllableRepository
		value   string
		wantErr error
	}{
		{
			name:    "invalid value",
			repo:    &controllableRepository{getConfig: config},
			value:   string(make([]byte, 10001)),
			wantErr: model.ErrInvalidValue,
		},
		{
			name:    "repository update error",
			repo:    &controllableRepository{getConfig: config, updateErr: errors.New("update failed")},
			value:   "new",
			wantErr: errors.New("update failed"),
		},
		{
			name:    "repository update not found",
			repo:    &controllableRepository{getConfig: config, updateErr: repository.ErrConfigNotFound},
			value:   "new",
			wantErr: ErrConfigNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigService(tt.repo).UpdateConfig("prod", "key", tt.value)
			if err == nil {
				t.Fatal("expected error")
			}
			if tt.wantErr == model.ErrInvalidValue || tt.wantErr == ErrConfigNotFound {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("UpdateConfig() error = %v, want %v", err, tt.wantErr)
				}
				if tt.wantErr == model.ErrInvalidValue && tt.repo.updated != nil {
					t.Fatal("repository Update should not be called after validation error")
				}
				return
			}
			if err.Error() != tt.wantErr.Error() {
				t.Fatalf("UpdateConfig() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigService_UpdateConfigReturnsGetDatabaseError(t *testing.T) {
	wantErr := errors.New("select failed")
	repo := &controllableRepository{getErr: wantErr}

	err := NewConfigService(repo).UpdateConfig("prod", "key", "value")
	if !errors.Is(err, wantErr) {
		t.Fatalf("UpdateConfig() error = %v, want %v", err, wantErr)
	}
	if repo.updated != nil {
		t.Fatal("repository Update should not be called after get error")
	}
}

func TestConfigService_DeleteConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		repo    *controllableRepository
		wantErr error
	}{
		{
			name:    "exists error",
			repo:    &controllableRepository{existsErr: errors.New("exists failed")},
			wantErr: errors.New("exists failed"),
		},
		{
			name:    "delete not found race",
			repo:    &controllableRepository{exists: true, deleteErr: repository.ErrConfigNotFound},
			wantErr: ErrConfigNotFound,
		},
		{
			name:    "delete database error",
			repo:    &controllableRepository{exists: true, deleteErr: errors.New("delete failed")},
			wantErr: errors.New("delete failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigService(tt.repo).DeleteConfig("prod", "key")
			if err == nil {
				t.Fatal("expected error")
			}
			if tt.wantErr == ErrConfigNotFound {
				if !errors.Is(err, ErrConfigNotFound) {
					t.Fatalf("DeleteConfig() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err.Error() != tt.wantErr.Error() {
				t.Fatalf("DeleteConfig() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
