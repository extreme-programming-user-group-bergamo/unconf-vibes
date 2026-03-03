package api

import (
	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/api/middleware"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())

	router.GET("/health", handlers.HealthHandler)

	return router
}
