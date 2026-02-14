package service

import (
	"config-service/internal/model"
	"errors"
	"testing"
)

type mockRepository struct {
	configs map[string]*model.Config
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		configs: make(map[string]*model.Config),
	}
}

func (m *mockRepository) Create(config *model.Config) error {
	key := config.Environment + ":" + config.Key
	if _, exists := m.configs[key]; exists {
		return errors.New("already exists")
	}
	m.configs[key] = config
	return nil
}

func (m *mockRepository) Get(environment, key string) (*model.Config, error) {
	lookupKey := environment + ":" + key
	config, exists := m.configs[lookupKey]
	if !exists {
		return nil, errors.New("not found")
	}
	return config, nil
}

func (m *mockRepository) GetAll(environment string) ([]*model.Config, error) {
	var result []*model.Config
	for key, config := range m.configs {
		if len(key) > len(environment)+1 && key[:len(environment)] == environment {
			result = append(result, config)
		}
	}
	return result, nil
}

func (m *mockRepository) Update(config *model.Config) error {
	key := config.Environment + ":" + config.Key
	if _, exists := m.configs[key]; !exists {
		return errors.New("not found")
	}
	m.configs[key] = config
	return nil
}

func (m *mockRepository) Delete(environment, key string) error {
	lookupKey := environment + ":" + key
	if _, exists := m.configs[lookupKey]; !exists {
		return errors.New("not found")
	}
	delete(m.configs, lookupKey)
	return nil
}

func (m *mockRepository) Exists(environment, key string) (bool, error) {
	lookupKey := environment + ":" + key
	_, exists := m.configs[lookupKey]
	return exists, nil
}

func TestConfigService_CreateConfig(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		key         string
		value       string
		setup       func(*mockRepository)
		wantErr     bool
		errType     error
	}{
		{
			name:        "successful create",
			environment: "prod",
			key:         "key1",
			value:       "value1",
			setup:       func(*mockRepository) {},
			wantErr:     false,
		},
		{
			name:        "config already exists",
			environment: "prod",
			key:         "key1",
			value:       "value1",
			setup: func(m *mockRepository) {
				config, _ := model.NewConfig("prod", "key1", "old_value")
				m.configs["prod:key1"] = config
			},
			wantErr: true,
			errType: ErrConfigExists,
		},
		{
			name:        "invalid environment",
			environment: "",
			key:         "key1",
			value:       "value1",
			setup:       func(*mockRepository) {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			tt.setup(repo)
			svc := NewConfigService(repo)

			err := svc.CreateConfig(tt.environment, tt.key, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("expected error %v, got %v", tt.errType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigService_GetConfig(t *testing.T) {
	repo := newMockRepository()
	config, _ := model.NewConfig("prod", "key1", "value1")
	repo.configs["prod:key1"] = config

	svc := NewConfigService(repo)

	tests := []struct {
		name        string
		environment string
		key         string
		wantErr     bool
	}{
		{
			name:        "successful get",
			environment: "prod",
			key:         "key1",
			wantErr:     false,
		},
		{
			name:        "not found",
			environment: "prod",
			key:         "key2",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := svc.GetConfig(tt.environment, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if !errors.Is(err, ErrConfigNotFound) {
					t.Errorf("expected ErrConfigNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if config == nil {
					t.Errorf("expected config but got nil")
				}
			}
		})
	}
}

func TestConfigService_UpdateConfig(t *testing.T) {
	repo := newMockRepository()
	config, _ := model.NewConfig("prod", "key1", "old_value")
	repo.configs["prod:key1"] = config

	svc := NewConfigService(repo)

	tests := []struct {
		name        string
		environment string
		key         string
		value       string
		wantErr     bool
	}{
		{
			name:        "successful update",
			environment: "prod",
			key:         "key1",
			value:       "new_value",
			wantErr:     false,
		},
		{
			name:        "not found",
			environment: "prod",
			key:         "key2",
			value:       "value",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateConfig(tt.environment, tt.key, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigService_DeleteConfig(t *testing.T) {
	repo := newMockRepository()
	config, _ := model.NewConfig("prod", "key1", "value1")
	repo.configs["prod:key1"] = config

	svc := NewConfigService(repo)

	tests := []struct {
		name        string
		environment string
		key         string
		wantErr     bool
	}{
		{
			name:        "successful delete",
			environment: "prod",
			key:         "key1",
			wantErr:     false,
		},
		{
			name:        "not found",
			environment: "prod",
			key:         "key2",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteConfig(tt.environment, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
