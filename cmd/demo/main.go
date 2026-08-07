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

	// The demo site keeps its fixtures in memory. It exists to have a page with
	// data-onboarding-id attributes to point at; giving it a database would add
	// a moving part that demonstrates nothing.
	listingDomain := domains.NewListingDomain(repositories.NewListingRepository())
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
