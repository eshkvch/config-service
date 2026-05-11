package di

import (
	"config-service/backend/config"
	"config-service/backend/internal/handler"
	"config-service/backend/internal/infrastructure/database"
	"config-service/backend/internal/repository"
	"config-service/backend/internal/service"
	"config-service/backend/pkg/metrics"
	"config-service/backend/pkg/server"
	"context"

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
			server.NewServer,
			metrics.NewMetrics,
		),
		fx.Invoke(registerLifecycle),
	)
}

func provideDatabaseConnection(cfg *config.Config) (database.Connection, error) {
	return database.NewPostgresConnection(cfg.Database.DSN)
}

func provideConfigRepository(conn database.Connection, m *metrics.Metrics) (repository.ConfigRepository, error) {
	return database.NewPostgresRepository(conn.GetDB(), m)
}

func provideConfigService(repo repository.ConfigRepository) service.ConfigService {
	return service.NewConfigService(repo)
}

func provideConfigHandler(svc service.ConfigService) *handler.ConfigHandler {
	return handler.NewConfigHandler(svc)
}

func registerLifecycle(
	lc fx.Lifecycle,
	server *server.Server,
	conn database.Connection,
	logger *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			server.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.GracefulShutdown(5)
			logger.Info("closing database connection")
			if err := conn.Close(); err != nil {
				logger.Error("failed to close database connection", zap.Error(err))
				return err
			}
			return nil
		},
	})
}
