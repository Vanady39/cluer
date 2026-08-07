//go:build debug
// +build debug

package server

import (
	_ "github.com/Vanady39/cluer/docs"
	"github.com/gin-gonic/gin"
	ginSwaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Gin-template
//	@version		0.0.1
//	@description	This is a sample server caller server.

//	@host		localhost:8080
//	@BasePath	/v1

func AddDocsForDebugVersion(router *gin.Engine) {
	gin.SetMode(gin.DebugMode)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(ginSwaggerFiles.Handler))
}
