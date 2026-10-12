package internal

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vlahanam/terra-clone/middleware"
)

func SetupRouter(c *AppConfig, accessWriter io.Writer) *gin.Engine {
	if c.AppEnv == "staging" || c.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger(accessWriter), gin.Recovery())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return r
}