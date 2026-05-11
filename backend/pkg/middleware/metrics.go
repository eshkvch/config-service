package middleware

import (
	"config-service/backend/pkg/metrics"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type MetricsMiddleware struct {
	metrics *metrics.Metrics
}

func NewMetricsMiddleware(m *metrics.Metrics) *MetricsMiddleware {
	return &MetricsMiddleware{metrics: m}
}

func (mw *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("middleware metrics")
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		mw.metrics.HTTPRequestsTotal.
			WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(rw.status)).
			Inc()

		mw.metrics.HTTPRequestDuration.
			WithLabelValues(r.URL.Path, r.Method).
			Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
