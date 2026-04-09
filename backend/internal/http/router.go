package http

import (
	"github.com/gin-gonic/gin"

	"bet/backend/internal/auth"
	"bet/backend/internal/config"
	"bet/backend/internal/http/handlers"
	"bet/backend/internal/http/middleware"
	"bet/backend/internal/notifications"
)

func NewRouter(cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	authService := auth.NewService()
	emailSender := notifications.NewSenderFromConfig(cfg)
	authHandler := handlers.NewAuthHandler(
		authService,
		emailSender,
		cfg.AuthJWTSecret,
		cfg.AuthTokenTTL,
		cfg.EmailVerifyBaseURL,
	)

	router.GET("/health", handlers.Health)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", handlers.Health)

		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.GET("/verify-email", authHandler.VerifyEmail)
		}

		meGroup := v1.Group("/me")
		meGroup.Use(middleware.RequireAuth(cfg.AuthJWTSecret))
		{
			meGroup.GET("", authHandler.Me)
		}

		moderationGroup := v1.Group("/moderation")
		moderationGroup.Use(middleware.RequireRoles(cfg.AuthJWTSecret, "moderator", "admin"))
		{
			moderationGroup.GET("/health", handlers.Health)
		}
	}

	return router
}
