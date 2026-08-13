package server

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/config"
	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/server/middlewares"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type DemoCreateStruct struct {
	Logger            *zerolog.Logger
	OIDCVerifier      *oidc.IDTokenVerifier
	ListingController controllers.ListingControllerInterface
	UserController    controllers.UserControllerInterface
}

func NewDemoServer(cfg *config.ServerConfig, createStruct *DemoCreateStruct) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		gin.Logger(),
		gin.Recovery(),
		middlewares.RuntimeCORS(),
		middlewares.ErrorHandler(createStruct.Logger),
	)

	v1 := router.Group("/v1")
	{
		v1.GET("/listings", createStruct.ListingController.GetListings)
		v1.GET("/users/me",
			middlewares.OIDCAuth(createStruct.OIDCVerifier),
			createStruct.UserController.GetCurrentUser,
		)
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
