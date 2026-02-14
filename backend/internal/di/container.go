package di

import (
	"config-service/backend/config"
	"config-service/backend/internal/handler"
	"config-service/backend/internal/infrastructure/database"
	"config-service/backend/internal/repository"
	"config-service/backend/internal/service"
	"context"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewApp() *fx.App {
	return fx.New(
		fx.Provide(
			config.Load,
			zap.NewProduction,
			provideDatabaseConnection,
			provideConfigRepository,
			provideConfigService,
			provideConfigHandler,
			provideHTTPServer,
		),
		fx.Invoke(registerLifecycle),
	)
}

func provideDatabaseConnection(cfg *config.Config) (database.Connection, error) {
	return database.NewPostgresConnection(cfg.Database.DSN)
}

func provideConfigRepository(conn database.Connection) (repository.ConfigRepository, error) {
	return database.NewPostgresRepository(conn.GetDB())
}

func provideConfigService(repo repository.ConfigRepository) service.ConfigService {
	return service.NewConfigService(repo)
}

func provideConfigHandler(svc service.ConfigService) *handler.ConfigHandler {
	return handler.NewConfigHandler(svc)
}

func provideHTTPServer(cfg *config.Config, h *handler.ConfigHandler) *http.Server {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.Handle(
		"/swagger/",
		httpSwagger.Handler(
			httpSwagger.URL("/doc.json"),
		),
	)
	return &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: mux,
	}
}

func registerLifecycle(
	lc fx.Lifecycle,
	server *http.Server,
	conn database.Connection,
	logger *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting http server", zap.String("addr", server.Addr))
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down http server")
			if err := server.Shutdown(ctx); err != nil {
				logger.Error("failed to shutdown http server", zap.Error(err))
				return err
			}
			logger.Info("closing database connection")
			if err := conn.Close(); err != nil {
				logger.Error("failed to close database connection", zap.Error(err))
				return err
			}
			return nil
		},
	})
}
