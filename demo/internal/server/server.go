package server

import (
	"github.com/Vanady39/cluer/demo/internal/controllers"
	"github.com/Vanady39/cluer/platform/config"
	"github.com/Vanady39/cluer/platform/middlewares"
	"github.com/Vanady39/cluer/platform/serve"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type CreateStruct struct {
	Logger            *zerolog.Logger
	OIDCVerifier      *oidc.IDTokenVerifier
	ListingController controllers.ListingControllerInterface
	UserController    controllers.UserControllerInterface
}

func NewServer(cfg *config.ServerConfig, createStruct *CreateStruct) *serve.Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		gin.Logger(),
		gin.Recovery(),
		middlewares.RuntimeCORS(),
		middlewares.ErrorHandler(createStruct.Logger),
	)

	AddDocsForDebugVersion(router)

	v1 := router.Group("/v1")
	{
		v1.GET("/listings", createStruct.ListingController.GetListings)
		v1.GET("/users/me",
			middlewares.OIDCAuth(createStruct.OIDCVerifier),
			createStruct.UserController.GetCurrentUser,
		)
	}

	createStruct.Logger.Debug().Msg("Demo routes initialized")

	return serve.New(cfg, createStruct.Logger, router.Handler())
}
