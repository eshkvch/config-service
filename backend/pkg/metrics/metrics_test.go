package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetricsRegistersCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()
	oldRegisterer := prometheus.DefaultRegisterer
	oldGatherer := prometheus.DefaultGatherer
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = oldRegisterer
		prometheus.DefaultGatherer = oldGatherer
	})

	m := NewMetrics()
	if m.HTTPRequestsTotal == nil ||
		m.HTTPRequestDuration == nil ||
		m.DBQueriesTotal == nil ||
		m.DBQueryDuration == nil {
		t.Fatal("expected all metric collectors to be initialized")
	}

	m.HTTPRequestsTotal.WithLabelValues("/api/configs/prod", "GET", "200").Inc()
	m.HTTPRequestDuration.WithLabelValues("/api/configs/prod", "GET").Observe(0.01)
	m.DBQueriesTotal.WithLabelValues("get").Inc()
	m.DBQueryDuration.WithLabelValues("get").Observe(0.02)

	gathered, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}

	names := make(map[string]bool)
	for _, family := range gathered {
		names[family.GetName()] = true
	}
	for _, name := range []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"db_queries_total",
		"db_query_duration_seconds",
	} {
		if !names[name] {
			t.Fatalf("metric %q was not registered", name)
		}
	}
}
