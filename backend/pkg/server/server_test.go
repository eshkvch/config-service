package server

import (
	"config-service/backend/config"
	"config-service/backend/internal/handler"
	"config-service/backend/internal/model"
	"config-service/backend/pkg/metrics"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type serverStubService struct{}

func (serverStubService) CreateConfig(string, string, string) error {
	return nil
}

func (serverStubService) GetConfig(environment, key string) (*model.Config, error) {
	return &model.Config{Environment: environment, Key: key, Value: "value"}, nil
}

func (serverStubService) GetAllConfigs(environment string) ([]*model.Config, error) {
	return []*model.Config{{Environment: environment, Key: "key", Value: "value"}}, nil
}

func (serverStubService) UpdateConfig(string, string, string) error {
	return nil
}

func (serverStubService) DeleteConfig(string, string) error {
	return nil
}

func serverTestMetrics() *metrics.Metrics {
	return &metrics.Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "server_test_http_requests_total"},
			[]string{"path", "method", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "server_test_http_request_duration_seconds"},
			[]string{"path", "method"},
		),
		DBQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "server_test_db_queries_total"},
			[]string{"operation"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "server_test_db_query_duration_seconds"},
			[]string{"operation"},
		),
	}
}

func TestNewServer(t *testing.T) {
	cfg := &config.Config{HTTP: config.HTTPConfig{Port: "18080"}}
	h := handler.NewConfigHandler(serverStubService{})

	srv := NewServer(cfg, h, serverTestMetrics())
	if srv == nil || srv.httpServer == nil {
		t.Fatal("server was not initialized")
	}
	if srv.httpServer.Addr != ":18080" {
		t.Fatalf("server addr = %q", srv.httpServer.Addr)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	srv.httpServer.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestProvideHTTPServerHandlesAPIRequest(t *testing.T) {
	cfg := &config.Config{HTTP: config.HTTPConfig{Port: "8081"}}
	h := handler.NewConfigHandler(serverStubService{})
	httpServer := provideHTTPServer(cfg, h, serverTestMetrics())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/configs/prod/key", nil)

	httpServer.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestServerStartCanBeShutdown(t *testing.T) {
	srv := &Server{
		httpServer: &http.Server{
			Addr:    "127.0.0.1:0",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		},
	}

	done := srv.Start()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.httpServer.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
}
