package http

import (
	"github.com/gin-gonic/gin"

	"bet/backend/internal/http/handlers"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	router.GET("/health", handlers.Health)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", handlers.Health)
	}

	return router
}
