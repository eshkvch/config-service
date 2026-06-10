package handler

import (
	"config-service/backend/internal/model"
	"config-service/backend/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubConfigService struct {
	createFunc func(environment, key, value string) error
	getFunc    func(environment, key string) (*model.Config, error)
	getAllFunc func(environment string) ([]*model.Config, error)
	updateFunc func(environment, key, value string) error
	deleteFunc func(environment, key string) error
}

func (s stubConfigService) CreateConfig(environment, key, value string) error {
	if s.createFunc != nil {
		return s.createFunc(environment, key, value)
	}
	return nil
}

func (s stubConfigService) GetConfig(environment, key string) (*model.Config, error) {
	if s.getFunc != nil {
		return s.getFunc(environment, key)
	}
	return &model.Config{Environment: environment, Key: key, Value: "value"}, nil
}

func (s stubConfigService) GetAllConfigs(environment string) ([]*model.Config, error) {
	if s.getAllFunc != nil {
		return s.getAllFunc(environment)
	}
	return []*model.Config{{Environment: environment, Key: "key", Value: "value"}}, nil
}

func (s stubConfigService) UpdateConfig(environment, key, value string) error {
	if s.updateFunc != nil {
		return s.updateFunc(environment, key, value)
	}
	return nil
}

func (s stubConfigService) DeleteConfig(environment, key string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(environment, key)
	}
	return nil
}

func TestConfigHandler_RegisterRoutesHealthAndDocs(t *testing.T) {
	mux := http.NewServeMux()
	NewConfigHandler(stubConfigService{}).RegisterRoutes(mux)

	tests := []struct {
		name        string
		method      string
		path        string
		wantStatus  int
		wantBody    string
		contentType string
	}{
		{
			name:        "health",
			method:      http.MethodGet,
			path:        "/health",
			wantStatus:  http.StatusOK,
			wantBody:    `"status":"ok"`,
			contentType: "application/json",
		},
		{
			name:       "health method not allowed",
			method:     http.MethodPost,
			path:       "/health",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "method not allowed",
		},
		{
			name:        "swagger json",
			method:      http.MethodGet,
			path:        "/doc.json",
			wantStatus:  http.StatusOK,
			wantBody:    `"openapi":"3.0.0"`,
			contentType: "application/json",
		},
		{
			name:        "swagger yaml",
			method:      http.MethodGet,
			path:        "/doc.yaml",
			wantStatus:  http.StatusOK,
			wantBody:    "openapi: 3.0.0",
			contentType: "text/yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%q", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want substring %q", rr.Body.String(), tt.wantBody)
			}
			if tt.contentType != "" && !strings.Contains(rr.Header().Get("Content-Type"), tt.contentType) {
				t.Fatalf("Content-Type = %q, want %q", rr.Header().Get("Content-Type"), tt.contentType)
			}
		})
	}
}

func TestConfigHandler_HandleConfigsRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		service    stubConfigService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing environment",
			method:     http.MethodGet,
			path:       "/api/configs/",
			wantStatus: http.StatusBadRequest,
			wantBody:   "environment is required",
		},
		{
			name:       "environment method not allowed",
			method:     http.MethodPost,
			path:       "/api/configs/prod",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "method not allowed",
		},
		{
			name:       "get all configs",
			method:     http.MethodGet,
			path:       "/api/configs/prod",
			wantStatus: http.StatusOK,
			wantBody:   `"key":"key"`,
		},
		{
			name:   "get all service error",
			method: http.MethodGet,
			path:   "/api/configs/prod",
			service: stubConfigService{
				getAllFunc: func(string) ([]*model.Config, error) {
					return nil, errors.New("db down")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "get config",
			method:     http.MethodGet,
			path:       "/api/configs/prod/key",
			wantStatus: http.StatusOK,
			wantBody:   `"value":"value"`,
		},
		{
			name:   "get not found",
			method: http.MethodGet,
			path:   "/api/configs/prod/missing",
			service: stubConfigService{
				getFunc: func(string, string) (*model.Config, error) {
					return nil, service.ErrConfigNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "config not found",
		},
		{
			name:       "create config",
			method:     http.MethodPost,
			path:       "/api/configs/prod/key",
			body:       `{"value":"created"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "create invalid json",
			method:     http.MethodPost,
			path:       "/api/configs/prod/key",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid json",
		},
		{
			name:   "create exists",
			method: http.MethodPost,
			path:   "/api/configs/prod/key",
			body:   `{"value":"created"}`,
			service: stubConfigService{
				createFunc: func(string, string, string) error {
					return service.ErrConfigExists
				},
			},
			wantStatus: http.StatusConflict,
			wantBody:   "config already exists",
		},
		{
			name:       "update config",
			method:     http.MethodPut,
			path:       "/api/configs/prod/key",
			body:       `{"value":"updated"}`,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "update invalid json",
			method:     http.MethodPut,
			path:       "/api/configs/prod/key",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid json",
		},
		{
			name:   "update service error",
			method: http.MethodPut,
			path:   "/api/configs/prod/key",
			body:   `{"value":"updated"}`,
			service: stubConfigService{
				updateFunc: func(string, string, string) error {
					return errors.New("validation")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "delete config",
			method:     http.MethodDelete,
			path:       "/api/configs/prod/key",
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "delete not found",
			method: http.MethodDelete,
			path:   "/api/configs/prod/key",
			service: stubConfigService{
				deleteFunc: func(string, string) error {
					return service.ErrConfigNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "config not found",
		},
		{
			name:       "unsupported item method",
			method:     http.MethodPatch,
			path:       "/api/configs/prod/key",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "method not allowed",
		},
		{
			name:       "invalid path",
			method:     http.MethodGet,
			path:       "/api/configs/prod/key/extra",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewConfigHandler(tt.service)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))

			h.handleConfigs(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%q", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want substring %q", rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestConfigHandler_CreateConfigHelper(t *testing.T) {
	var gotEnvironment, gotKey, gotValue string
	h := NewConfigHandler(stubConfigService{
		createFunc: func(environment, key, value string) error {
			gotEnvironment = environment
			gotKey = key
			gotValue = value
			return nil
		},
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/configs/prod/key", strings.NewReader(`{"value":"created"}`))

	h.createConfig(rr, req, "prod", "key")

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if gotEnvironment != "prod" || gotKey != "key" || gotValue != "created" {
		t.Fatalf("CreateConfig called with %q, %q, %q", gotEnvironment, gotKey, gotValue)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/configs/prod/key", strings.NewReader(`{`))
	h.createConfig(rr, req, "prod", "key")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid json status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	h = NewConfigHandler(stubConfigService{
		createFunc: func(string, string, string) error {
			return service.ErrConfigExists
		},
	})
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/configs/prod/key", strings.NewReader(`{"value":"created"}`))
	h.createConfig(rr, req, "prod", "key")
	if rr.Code != http.StatusConflict {
		t.Fatalf("service error status = %d, want %d", rr.Code, http.StatusConflict)
	}
}

func TestConfigHandler_JSONResponseShape(t *testing.T) {
	h := NewConfigHandler(stubConfigService{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/configs/prod/key", nil)

	h.getConfig(rr, req, "prod", "key")

	var got model.Config
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Environment != "prod" || got.Key != "key" || got.Value != "value" {
		t.Fatalf("decoded config = %#v", got)
	}
}
