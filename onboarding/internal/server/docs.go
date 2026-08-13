//go:build !debug
// +build !debug

package server

import "github.com/gin-gonic/gin"

func AddDocsForDebugVersion(router *gin.Engine) {}
