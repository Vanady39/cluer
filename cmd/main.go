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

//	@title			Cluer
//	@version		0.0.1
//	@description	Onboarding tours and listings API.

// @host		localhost:8080
// @BasePath	/v1

func main() {
	// Read config
	cfg := config.Load()

	// Setup logger
	defaultLogger := logger.New(cfg.Logger.Default.Level)
	defaultLogger.Info().Msg("Logger setup successfully")

	// --------------
	// -- LISTINGS --
	// --------------

	listingRepo := repositories.NewListingRepository()
	listingDomain := domains.NewListingDomain(listingRepo)
	listingController := controllers.NewListingController(listingDomain)

	// -----------
	// -- USERS --
	// -----------

	userRepo := repositories.NewUserRepository()
	userDomain := domains.NewUserDomain(userRepo)
	userController := controllers.NewUserController(userDomain)

	// ----------------
	// -- ONBOARDING --
	// ----------------

	// Tours and hints share both repositories: creating a hint has to check the
	// status of its parent tour, and listing published tours pulls in their hints.
	tourRepo := repositories.NewTourRepository()
	hintRepo := repositories.NewHintRepository()

	tourDomain := domains.NewTourDomain(tourRepo, hintRepo)
	hintDomain := domains.NewHintDomain(tourRepo, hintRepo)

	tourController := controllers.NewTourController(tourDomain)
	hintController := controllers.NewHintController(hintDomain)

	// Setup server
	srvLogger := logger.New(cfg.GetLoggerConfig("server").Level)
	srvCreateStruct := &server.CreateStruct{
		Logger:            srvLogger,
		ListingController: listingController,
		UserController:    userController,
		TourController:    tourController,
		HintController:    hintController,
	}
	srv := server.NewServer(cfg.ServerConfig, srvCreateStruct)
	defaultLogger.Debug().Msg("Server created successfully")

	// Launch application
	go func() {
		if err := srv.Start(); err != nil {
			defaultLogger.Fatal().Err(err).Msg("Application failed to run")
		}
	}()

	// Graceful shutdown
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
