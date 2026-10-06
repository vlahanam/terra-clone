package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(c *AppConfig) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return r
}