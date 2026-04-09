package http

import (
	"github.com/gin-gonic/gin"

	"bet/backend/internal/auth"
	"bet/backend/internal/bets"
	"bet/backend/internal/config"
	"bet/backend/internal/events"
	"bet/backend/internal/http/handlers"
	"bet/backend/internal/http/middleware"
	"bet/backend/internal/notifications"
	"bet/backend/internal/wallet"
)

func NewRouter(cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	authService := auth.NewService()
	eventsService := events.NewService()
	walletService := wallet.NewService(1000)
	betsService := bets.NewService(eventsService, walletService)
	emailSender := notifications.NewSenderFromConfig(cfg)
	authHandler := handlers.NewAuthHandler(
		authService,
		emailSender,
		cfg.AuthJWTSecret,
		cfg.AuthTokenTTL,
		cfg.EmailVerifyBaseURL,
	)
	eventsHandler := handlers.NewEventsHandler(eventsService, betsService)
	walletHandler := handlers.NewWalletHandler(walletService)
	betsHandler := handlers.NewBetsHandler(betsService)

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

		walletGroup := v1.Group("/wallet")
		walletGroup.Use(middleware.RequireAuth(cfg.AuthJWTSecret))
		{
			walletGroup.GET("", walletHandler.GetWallet)
			walletGroup.GET("/transactions", walletHandler.ListTransactions)
		}

		betsGroup := v1.Group("/bets")
		betsGroup.Use(middleware.RequireAuth(cfg.AuthJWTSecret))
		{
			betsGroup.POST("", betsHandler.PlaceBet)
			betsGroup.GET("/my", betsHandler.ListMyBets)
		}

		eventsGroup := v1.Group("/events")
		{
			eventsGroup.GET("", eventsHandler.ListEvents)
			eventsGroup.GET("/:id", eventsHandler.GetEvent)

			eventsAuthGroup := eventsGroup.Group("")
			eventsAuthGroup.Use(middleware.RequireAuth(cfg.AuthJWTSecret))
			{
				eventsAuthGroup.POST("", eventsHandler.CreateEvent)
			}
		}

		moderationGroup := v1.Group("/moderation")
		moderationGroup.Use(middleware.RequireRoles(cfg.AuthJWTSecret, "moderator", "admin"))
		{
			moderationGroup.GET("/health", handlers.Health)
			moderationGroup.GET("/events", eventsHandler.ListModerationEvents)
			moderationGroup.POST("/events/:id/approve", eventsHandler.ApproveEvent)
			moderationGroup.POST("/events/:id/reject", eventsHandler.RejectEvent)
		}

		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.RequireRoles(cfg.AuthJWTSecret, "admin"))
		{
			adminGroup.POST("/events/:id/settle", eventsHandler.SettleEvent)
		}
	}

	return router
}
