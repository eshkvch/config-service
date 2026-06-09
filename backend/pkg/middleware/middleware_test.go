package middleware

import (
	"config-service/backend/pkg/metrics"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func testMetrics() *metrics.Metrics {
	return &metrics.Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "middleware_test_http_requests_total"},
			[]string{"path", "method", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "middleware_test_http_request_duration_seconds"},
			[]string{"path", "method"},
		),
		DBQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "middleware_test_db_queries_total"},
			[]string{"operation"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "middleware_test_db_query_duration_seconds"},
			[]string{"operation"},
		),
	}
}

func TestLogMiddleware(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/path", nil)

	LogMiddleware(next).ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}
}

func TestMetricsMiddlewareRecordsRequestStatus(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("body"))
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/configs/prod/key", nil)

	NewMetricsMiddleware(testMetrics()).Handler(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusTeapot)
	}
	if rr.Body.String() != "body" {
		t.Fatalf("body = %q", rr.Body.String())
	}
}

func TestMetricsMiddlewareSkipsConfiguredPaths(t *testing.T) {
	for _, path := range []string{"/metrics", "/health", "/swagger/index.html", "/doc.json", "/doc.yaml"} {
		t.Run(path, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)

			NewMetricsMiddleware(nil).Handler(next).ServeHTTP(rr, req)

			if !nextCalled {
				t.Fatal("next handler was not called")
			}
			if rr.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
			}
		})
	}
}

func TestShouldSkipMetrics(t *testing.T) {
	tests := map[string]bool{
		"/metrics":              true,
		"/health":               true,
		"/swagger/doc":          true,
		"/doc.json":             true,
		"/api/configs/prod/key": false,
		"/api/configs/prod":     false,
	}

	for path, want := range tests {
		if got := shouldSkipMetrics(path); got != want {
			t.Fatalf("shouldSkipMetrics(%q) = %v, want %v", path, got, want)
		}
	}
}
