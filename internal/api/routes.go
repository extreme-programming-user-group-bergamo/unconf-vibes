package api

import (
	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/api/middleware"
)

func NewRouter(
	authHandler *handlers.AuthHandler,
	tokenValidator middleware.TokenValidator,
	userHandler *handlers.UserHandler,
	confHandler *handlers.ConferenceHandler,
	roomHandler *handlers.RoomHandler,
	attendeeHandler *handlers.AttendeeHandler,
	requestHandler *handlers.RequestHandler,
) *gin.Engine {
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

	if roomHandler != nil {
		router.GET("/conferences/:slug/rooms", roomHandler.ListByConference)
	}

	if tokenValidator != nil {
		protected := router.Group("")
		protected.Use(middleware.AuthMiddleware(tokenValidator))

		protected.POST("/auth/revoke", authHandler.Revoke)

		if userHandler != nil {
			protected.GET("/users/me", userHandler.GetMe)
			protected.PUT("/users/me", userHandler.UpdateMe)
		}

		if attendeeHandler != nil {
			protected.GET("/conferences/:slug/attendees", attendeeHandler.ListByConference)
		}

		if requestHandler != nil {
			protected.POST("/requests", requestHandler.Create)
			protected.GET("/requests", requestHandler.List)
			protected.PUT("/requests/:id/accept", requestHandler.Accept)
			protected.PUT("/requests/:id/decline", requestHandler.Decline)
		}
	}

	return router
}
