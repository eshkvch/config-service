package model

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		key         string
		value       string
		wantErr     bool
		errType     error
	}{
		{
			name:        "valid config",
			environment: "production",
			key:         "database_url",
			value:       "postgres://localhost:5432/db",
			wantErr:     false,
		},
		{
			name:        "empty environment",
			environment: "",
			key:         "key",
			value:       "value",
			wantErr:     true,
			errType:     ErrInvalidEnvironment,
		},
		{
			name:        "empty key",
			environment: "prod",
			key:         "",
			value:       "value",
			wantErr:     true,
			errType:     ErrInvalidKey,
		},
		{
			name:        "too long environment",
			environment: string(make([]byte, 101)),
			key:         "key",
			value:       "value",
			wantErr:     true,
			errType:     ErrInvalidEnvironment,
		},
		{
			name:        "too long key",
			environment: "prod",
			key:         string(make([]byte, 256)),
			value:       "value",
			wantErr:     true,
			errType:     ErrInvalidKey,
		},
		{
			name:        "too long value",
			environment: "prod",
			key:         "key",
			value:       string(make([]byte, 10001)),
			wantErr:     true,
			errType:     ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.environment, tt.key, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err != tt.errType {
					t.Errorf("expected error %v, got %v", tt.errType, err)
				}
				if config != nil {
					t.Errorf("expected nil config on error, got %v", config)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if config == nil {
					t.Errorf("expected config but got nil")
					return
				}
				if config.Environment != tt.environment {
					t.Errorf("expected environment %s, got %s", tt.environment, config.Environment)
				}
				if config.Key != tt.key {
					t.Errorf("expected key %s, got %s", tt.key, config.Key)
				}
				if config.Value != tt.value {
					t.Errorf("expected value %s, got %s", tt.value, config.Value)
				}
			}
		})
	}
}

func TestConfig_UpdateValue(t *testing.T) {
	config, err := NewConfig("prod", "key", "old_value")
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid update",
			value:   "new_value",
			wantErr: false,
		},
		{
			name:    "too long value",
			value:   string(make([]byte, 10001)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldValue := config.Value
			err := config.UpdateValue(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if config.Value != oldValue {
					t.Errorf("value should not change on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if config.Value != tt.value {
					t.Errorf("expected value %s, got %s", tt.value, config.Value)
				}
			}
		})
	}
}
