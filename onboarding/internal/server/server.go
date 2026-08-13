package server

import (
	"github.com/Vanady39/cluer/onboarding/internal/controllers"
	"github.com/Vanady39/cluer/onboarding/internal/domains"
	appmw "github.com/Vanady39/cluer/onboarding/internal/middlewares"
	"github.com/Vanady39/cluer/platform/config"
	"github.com/Vanady39/cluer/platform/middlewares"
	"github.com/Vanady39/cluer/platform/serve"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type CreateStruct struct {
	Logger            *zerolog.Logger
	RuntimeDomain     domains.RuntimeDomainInterface
	OIDCVerifier      *oidc.IDTokenVerifier
	AdminGroups       []string
	TourController    controllers.TourControllerInterface
	HintController    controllers.HintControllerInterface
	RuntimeController controllers.RuntimeControllerInterface
	MeController      controllers.MeControllerInterface
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
		v1.GET("/health", createStruct.RuntimeController.Health)

		runtime := v1.Group("")
		runtime.Use(appmw.AppKeyAuth(createStruct.RuntimeDomain))
		{
			runtime.POST("/resolve", createStruct.RuntimeController.Resolve)
			runtime.POST("/events", createStruct.RuntimeController.Ingest)
		}

		admin := v1.Group("")
		admin.Use(
			middlewares.OIDCAuth(createStruct.OIDCVerifier),
			middlewares.RequireGroups(createStruct.AdminGroups),
		)
		{
			admin.GET("/users/me", createStruct.MeController.GetCurrentAdmin)

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

	return serve.New(cfg, createStruct.Logger, router.Handler())
}
