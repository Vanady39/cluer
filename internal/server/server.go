package server

import (
	"context"
	"net/http"

	"github.com/Vanady39/cluer/internal/config"
	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/server/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type (
	// ServerInterface defines interface of server of this project
	ServerInterface interface {
		Start() error
		Shutdown(ctx context.Context) error
	}

	// Server responsible for up HTTP server
	Server struct {
		logger     *zerolog.Logger
		httpServer *http.Server
	}

	CreateStruct struct {
		Logger            *zerolog.Logger
		RuntimeDomain     domains.RuntimeDomainInterface
		TourController    controllers.TourControllerInterface
		HintController    controllers.HintControllerInterface
		RuntimeController controllers.RuntimeControllerInterface
	}
)

// NewServer wires the routes.
//
// The admin group is currently open: authentication is deliberately deferred and
// everything acts as a single mocked user. The grouping below is kept anyway, so
// that switching it back on is one middleware line rather than a reshuffle of
// the route tree.
//
// This is fine for local development and is not fine anywhere else: /v1/tours
// can publish, roll back and archive, so an open deployment lets anyone switch
// off every tour in production.
func NewServer(cfg *config.ServerConfig, createStruct *CreateStruct) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
		middlewares.ErrorHandler(createStruct.Logger),
	)

	AddDocsForDebugVersion(router)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", createStruct.RuntimeController.Health)

		// ---- runtime: the SDK, authenticated by the app's public key ----
		runtime := v1.Group("")
		runtime.Use(
			middlewares.RuntimeCORS(),
			middlewares.AppKeyAuth(createStruct.RuntimeDomain),
		)
		{
			runtime.POST("/resolve", createStruct.RuntimeController.Resolve)
			runtime.POST("/events", createStruct.RuntimeController.Ingest)
			runtime.OPTIONS("/resolve", func(*gin.Context) {})
			runtime.OPTIONS("/events", func(*gin.Context) {})
		}

		// ---- admin: everything that can change what is live ----
		// TODO: re-enable middlewares.AdminAuth here once auth lands.
		admin := v1.Group("")
		{
			admin.POST("/apps", createStruct.RuntimeController.CreateApp)
			admin.GET("/apps", createStruct.RuntimeController.ListApps)

			admin.GET("/versions/:versionId", createStruct.TourController.GetVersion)

			tours := admin.Group("/tours")
			{
				tours.POST("", createStruct.TourController.Create)
				tours.GET("", createStruct.TourController.List)
				tours.GET("/published", createStruct.TourController.GetPublished)

				tourID := tours.Group("/:tourId")
				{
					tourID.GET("", createStruct.TourController.Card)
					tourID.PATCH("", createStruct.TourController.UpdateMeta)
					tourID.DELETE("", createStruct.TourController.Archive)

					tourID.GET("/versions", createStruct.TourController.ListVersions)
					tourID.POST("/draft", createStruct.TourController.CreateDraft)
					tourID.PATCH("/draft", createStruct.TourController.UpdateDraft)
					tourID.POST("/publish", createStruct.TourController.Publish)
					tourID.POST("/rollback", createStruct.TourController.Rollback)
					tourID.GET("/analytics", createStruct.RuntimeController.Analytics)

					hints := tourID.Group("/hints")
					{
						hints.POST("", createStruct.HintController.Create)
						hints.GET("", createStruct.HintController.List)
						hints.PUT("/order", createStruct.HintController.Reorder)
						hints.PATCH("/:hintId", createStruct.HintController.Update)
						hints.DELETE("/:hintId", createStruct.HintController.Delete)
					}
				}
			}
		}
	}

	createStruct.Logger.Debug().Msg("Routes initialized")

	return &Server{
		logger: createStruct.Logger,
		httpServer: &http.Server{
			Addr:    cfg.Address,
			Handler: router.Handler(),
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info().Str("address", s.httpServer.Addr).Msg("Starting server...")
	// ListenAndServe blocks until shutdown, so there is nothing to log on the way
	// out beyond a genuine failure.
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error().Err(err).Msg("Server failed to start")
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Server failed to shutdown")
		return err
	}
	return nil
}
