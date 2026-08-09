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

//	@title			Cluer Demo Site
//	@version		0.1.0
//	@description	Sample classifieds app the onboarding platform is demonstrated on.

// @host		localhost:8081
// @BasePath	/v1
func main() {
	cfg := config.Load()

	defaultLogger := logger.New(cfg.Logger.Default.Level)
	defaultLogger.Info().Msg("Logger setup successfully")

	if cfg.PostgresConfig == nil {
		defaultLogger.Fatal().Msg("postgres config is required for the demo service")
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), time.Minute)
	defer startupCancel()

	dsn := cfg.PostgresConfig.GetDSN()
	if err := storage.Migrate(dsn, defaultLogger); err != nil {
		defaultLogger.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Migrations failed")
	}

	pool, err := storage.Connect(startupCtx, dsn, defaultLogger)
	if err != nil {
		defaultLogger.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Database connection failed")
	}
	defer pool.Close()

	if err := storage.SeedDemoListings(startupCtx, pool); err != nil {
		defaultLogger.Fatal().Err(err).Msg("Demo listings seed failed")
	}

	listingDomain := domains.NewListingDomain(repositories.NewListingRepository(pool))
	userDomain := domains.NewUserDomain(repositories.NewUserRepository())

	srvLogger := logger.New(cfg.GetLoggerConfig("server").Level)
	srv := server.NewDemoServer(cfg.ServerConfig, &server.DemoCreateStruct{
		Logger:            srvLogger,
		ListingController: controllers.NewListingController(listingDomain),
		UserController:    controllers.NewUserController(userDomain),
	})

	go func() {
		if err := srv.Start(); err != nil {
			defaultLogger.Fatal().Err(err).Msg("Application failed to run")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		defaultLogger.Error().Err(err).Msg("Server shutdown")
	}

	defaultLogger.Info().Msg("Server exiting")
}
