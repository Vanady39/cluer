//go:build debug
// +build debug

package server

import (
	_ "github.com/Vanady39/cluer/demo/docs"
	"github.com/gin-gonic/gin"
	ginSwaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AddDocsForDebugVersion(router *gin.Engine) {
	gin.SetMode(gin.DebugMode)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(ginSwaggerFiles.Handler))
}
