package api

import (
	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/api/middleware"
)

func NewRouter(authHandler *handlers.AuthHandler, tokenValidator middleware.TokenValidator, userHandler *handlers.UserHandler, confHandler *handlers.ConferenceHandler) *gin.Engine {
	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())

	if authHandler == nil {
		authHandler = handlers.NewAuthHandler(nil)
	}

	router.GET("/health", handlers.HealthHandler)
	router.POST("/auth/device", authHandler.StartDeviceFlow)
	router.POST("/auth/token", authHandler.ExchangeDeviceCode)
	router.POST("/auth/refresh", authHandler.Refresh)

	if confHandler != nil {
		router.GET("/conferences", confHandler.List)
		router.GET("/conferences/:slug", confHandler.GetBySlug)
	}

	if tokenValidator != nil {
		protected := router.Group("")
		protected.Use(middleware.AuthMiddleware(tokenValidator))

		protected.POST("/auth/revoke", authHandler.Revoke)

		if userHandler != nil {
			protected.GET("/users/me", userHandler.GetMe)
			protected.PUT("/users/me", userHandler.UpdateMe)
		}
	}

	return router
}
