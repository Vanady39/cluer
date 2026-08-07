package server

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/config"
	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/server/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// DemoCreateStruct wires the sample classifieds site the onboarding is
// demonstrated on.
type DemoCreateStruct struct {
	Logger            *zerolog.Logger
	ListingController controllers.ListingControllerInterface
	UserController    controllers.UserControllerInterface
}

// NewDemoServer serves the test site.
//
// It is a separate binary from the onboarding platform on purpose. The platform
// is the product — a mechanism that attaches to any web app — and the
// classifieds site is one of its consumers. Keeping them in one process would
// make that boundary invisible and tempt exactly the coupling the design is
// meant to prevent.
func NewDemoServer(cfg *config.ServerConfig, createStruct *DemoCreateStruct) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		gin.Logger(),
		gin.Recovery(),
		middlewares.ErrorHandler(createStruct.Logger),
	)

	v1 := router.Group("/v1")
	{
		v1.GET("/listings", createStruct.ListingController.GetListings)
		v1.GET("/users/me", createStruct.UserController.GetCurrentUser)
	}

	createStruct.Logger.Debug().Msg("Demo routes initialized")

	return &Server{
		logger: createStruct.Logger,
		httpServer: &http.Server{
			Addr:    cfg.Address,
			Handler: router.Handler(),
		},
	}
}
