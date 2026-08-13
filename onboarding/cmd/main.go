package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vanady39/cluer/onboarding/internal/controllers"
	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/Vanady39/cluer/onboarding/internal/repositories"
	"github.com/Vanady39/cluer/onboarding/internal/server"
	"github.com/Vanady39/cluer/platform/config"
	"github.com/Vanady39/cluer/platform/logger"
	"github.com/Vanady39/cluer/platform/pg"
	"github.com/coreos/go-oidc/v3/oidc"
)

//	@title			Cluer Onboarding
//	@version		0.1.0
//	@description	Interactive onboarding platform: tours, immutable versions, resolve and analytics.

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				OIDC id token: "Bearer <id_token>".

// @host		localhost:8080
// @BasePath	/v1
func main() {
	cfg := config.Load()

	defaultLogger := logger.New(cfg.Logger.Default.Level)
	defaultLogger.Info().Msg("Logger setup successfully")

	if cfg.PostgresConfig == nil {
		defaultLogger.Fatal().Msg("postgres config is required for the onboarding service")
	}

	if cfg.OIDCConfig == nil {
		defaultLogger.Fatal().Msg("oidc config is required for the onboarding service")
	}

	// admin_groups is optional at the config level because cmd/demo shares this
	// struct and has no admin API to guard. Here it is not optional: an empty
	// list means RequireGroups can never match, so every administrator would be
	// refused with 403 while the service looks perfectly healthy. Refusing to
	// start says out loud what a silent deny-all would only whisper.
	if len(cfg.OIDCConfig.AdminGroups) == 0 {
		defaultLogger.Fatal().Msg("oidc.admin_groups must list at least one group, otherwise nobody can reach the admin API")
	}

	// Discovery is a synchronous network call, and the service is useless
	// without it: with no verifier every admin request would have to be
	// refused, so failing to start is the honest outcome.
	//
	// context.Background() and not the timed ctx below on purpose — the
	// provider keeps this context and reuses it to refresh the JWKS for the
	// lifetime of the process. A cancelled one would silently break key
	// rotation minutes after a successful start.
	provider, err := oidc.NewProvider(context.Background(), cfg.OIDCConfig.Issuer)
	if err != nil {
		defaultLogger.Fatal().Err(err).Str("issuer", cfg.OIDCConfig.Issuer).
			Msg("Failed to initialize OIDC provider")
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.OIDCConfig.ClientID})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dsn := cfg.PostgresConfig.GetDSN()
	pool, err := pg.Connect(ctx, dsn, defaultLogger)
	if err != nil {
		defaultLogger.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Database connection failed")
	}
	defer pool.Close()

	tourRepo := repositories.NewTourRepository(pool, defaultLogger)
	hintRepo := repositories.NewHintRepository(pool, defaultLogger)
	runtimeRepo := repositories.NewRuntimeRepository(pool, defaultLogger)

	tourDomain := domains.NewTourDomain(tourRepo, hintRepo)
	hintDomain := domains.NewHintDomain(tourRepo, hintRepo)
	runtimeDomain := domains.NewRuntimeDomain(tourRepo, hintRepo, runtimeRepo)

	srvLogger := logger.New(cfg.GetLoggerConfig("server").Level)
	srv := server.NewServer(cfg.ServerConfig, &server.CreateStruct{
		Logger:            srvLogger,
		RuntimeDomain:     runtimeDomain,
		OIDCVerifier:      verifier,
		AdminGroups:       cfg.OIDCConfig.AdminGroups,
		TourController:    controllers.NewTourController(tourDomain),
		HintController:    controllers.NewHintController(hintDomain),
		RuntimeController: controllers.NewRuntimeController(runtimeDomain),
		MeController:      controllers.NewMeController(),
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
