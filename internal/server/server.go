package server

import (
	"context"
	"net/http"

	"github.com/Vanady39/cluer/internal/config"
	"github.com/Vanady39/cluer/internal/controllers"
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
		ListingController controllers.ListingControllerInterface
		UserController    controllers.UserControllerInterface
		TourController    controllers.TourControllerInterface
		HintController    controllers.HintControllerInterface
	}
)

// NewServer creates new entity of Server
func NewServer(cfg *config.ServerConfig, createStruct *CreateStruct) *Server {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create a new Gin router instance
	router := gin.New()

	// Add middlewares
	router.Use(
		gin.Logger(),
		gin.Recovery(),
		middlewares.ErrorHandler(createStruct.Logger),
	)

	// If debug is enabled add swagger endpoint
	AddDocsForDebugVersion(router)

	// Setup routes
	v1 := router.Group("/v1")
	{
		v1.GET("/listings", createStruct.ListingController.GetListings)

		users := v1.Group("/users")
		{
			users.GET("/me", createStruct.UserController.GetCurrentUser)
		}

		tours := v1.Group("/tours")
		{
			tours.POST("", createStruct.TourController.Create)
			tours.GET("/published", createStruct.TourController.GetPublished)

			tourID := tours.Group("/:tourId")
			{
				tourID.POST("/hints", createStruct.HintController.Create)
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
