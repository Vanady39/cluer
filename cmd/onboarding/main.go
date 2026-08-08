package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vanady39/cluer/internal/config"
	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/logger"
	"github.com/Vanady39/cluer/internal/repositories"
	"github.com/Vanady39/cluer/internal/server"
	"github.com/Vanady39/cluer/internal/storage"
)

//	@title			Cluer Onboarding
//	@version		0.1.0
//	@description	Interactive onboarding platform: tours, immutable versions, resolve and analytics.

// @host		localhost:8080
// @BasePath	/v1
func main() {
	cfg := config.Load()

	defaultLogger := logger.New(cfg.Logger.Default.Level)
	defaultLogger.Info().Msg("Logger setup successfully")

	if cfg.PostgresConfig == nil {
		defaultLogger.Fatal().Msg("postgres config is required for the onboarding service")
	}

	defaultLogger.Warn().Msg("Admin endpoints are unauthenticated: acting as a single mocked user")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dsn := cfg.PostgresConfig.GetDSN()

	if err := storage.Migrate(dsn, defaultLogger); err != nil {
		defaultLogger.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Migrations failed")
	}

	pool, err := storage.Connect(ctx, dsn, defaultLogger)
	if err != nil {
		defaultLogger.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Database connection failed")
	}
	defer pool.Close()

	tourRepo := repositories.NewTourRepository(pool)
	hintRepo := repositories.NewHintRepository(pool)
	runtimeRepo := repositories.NewRuntimeRepository(pool)

	tourDomain := domains.NewTourDomain(tourRepo, hintRepo)
	hintDomain := domains.NewHintDomain(tourRepo, hintRepo)
	runtimeDomain := domains.NewRuntimeDomain(tourRepo, hintRepo, runtimeRepo)

	srvLogger := logger.New(cfg.GetLoggerConfig("server").Level)
	srv := server.NewServer(cfg.ServerConfig, &server.CreateStruct{
		Logger:            srvLogger,
		RuntimeDomain:     runtimeDomain,
		TourController:    controllers.NewTourController(tourDomain),
		HintController:    controllers.NewHintController(hintDomain),
		RuntimeController: controllers.NewRuntimeController(runtimeDomain),
	})
	defaultLogger.Debug().Msg("Server created successfully")

	go func() {
		if err := srv.Start(); err != nil {
			defaultLogger.Fatal().Err(err).Msg("Application failed to run")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		defaultLogger.Error().Err(err).Msg("Server shutdown")
	}

	defaultLogger.Info().Msg("Server exiting")
}
