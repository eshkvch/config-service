package server

import (
	"config-service/backend/config"
	"config-service/backend/internal/handler"
	"config-service/backend/pkg/middleware"
	"config-service/backend/pkg/metrics"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	httpServer *http.Server
}

func provideHTTPServer(
	cfg *config.Config,
	h *handler.ConfigHandler,
	m *metrics.Metrics,
) *http.Server {

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	mux.Handle(
		"/swagger/",
		httpSwagger.Handler(httpSwagger.URL("/doc.json")),
	)

	mux.Handle("/metrics", promhttp.Handler())

	metricsMw := middleware.NewMetricsMiddleware(m)

	var handler http.Handler = mux
	handler = metricsMw.Handler(handler)

	return &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: handler,
	}
}

func NewServer(
	cfg *config.Config,
	h *handler.ConfigHandler,
	m *metrics.Metrics,
) *Server {
	return &Server{
		httpServer: provideHTTPServer(cfg, h, m),
	}
}

func (s *Server) Start() <-chan struct{} {
	done := make(chan struct{})

	go func() {
		log.Println("HTTP server started on", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
		close(done)
	}()

	return done
}

func (s *Server) GracefulShutdown(timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown signal received, stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
