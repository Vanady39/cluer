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
	"github.com/coreos/go-oidc/v3/oidc"
)

//	@title			Cluer Demo Site
//	@version		0.1.0
//	@description	Sample classifieds app the onboarding platform is demonstrated on.

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				OIDC id token: "Bearer <id_token>".

// @host		localhost:8081
// @BasePath	/v1
func main() {
	cfg := config.Load()

	defaultLogger := logger.New(cfg.Logger.Default.Level)
	defaultLogger.Info().Msg("Logger setup successfully")

	if cfg.PostgresConfig == nil {
		defaultLogger.Fatal().Msg("postgres config is required for the demo service")
	}

	if cfg.OIDCConfig == nil {
		defaultLogger.Fatal().Msg("oidc config is required for the demo service")
	}

	// Same reasoning as in the onboarding service: discovery is a synchronous
	// network call and there is no verifier without it, so /users/me could only
	// ever answer 401. Refusing to start says that out loud instead of serving
	// a storefront where nobody can be logged in. Only the verifier is built
	// from the config: admin_groups belongs to the admin API and is never read
	// here, because on the storefront every authenticated visitor is a user.
	//
	// context.Background() and not a timed one on purpose: the provider keeps
	// this context to refresh the JWKS for the life of the process, and a
	// cancelled one would break key rotation long after a successful start.
	provider, err := oidc.NewProvider(context.Background(), cfg.OIDCConfig.Issuer)
	if err != nil {
		defaultLogger.Fatal().Err(err).Str("issuer", cfg.OIDCConfig.Issuer).
			Msg("Failed to initialize OIDC provider")
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.OIDCConfig.ClientID})

	startupCtx, startupCancel := context.WithTimeout(context.Background(), time.Minute)
	defer startupCancel()

	dsn := cfg.PostgresConfig.GetDSN()
	pool, err := storage.Connect(startupCtx, dsn, defaultLogger)
	if err != nil {
		defaultLogger.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Database connection failed")
	}
	defer pool.Close()

	if err := storage.SeedDemoListings(startupCtx, pool); err != nil {
		defaultLogger.Fatal().Err(err).Msg("Demo listings seed failed")
	}

	listingDomain := domains.NewListingDomain(repositories.NewListingRepository(pool))

	srvLogger := logger.New(cfg.GetLoggerConfig("server").Level)
	srv := server.NewDemoServer(cfg.ServerConfig, &server.DemoCreateStruct{
		Logger:            srvLogger,
		OIDCVerifier:      verifier,
		ListingController: controllers.NewListingController(listingDomain),
		UserController:    controllers.NewUserController(),
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
