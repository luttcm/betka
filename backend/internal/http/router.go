package http

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

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

	var db *sql.DB
	if cfg.DatabaseURL != "" {
		opened, err := sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			if cfg.RequirePostgres {
				panic(fmt.Sprintf("failed to initialize postgres driver: %v", err))
			}
			log.Printf("failed to initialize postgres driver, fallback to in-memory services: %v", err)
		} else {
			opened.SetConnMaxLifetime(30 * time.Minute)
			opened.SetMaxOpenConns(10)
			opened.SetMaxIdleConns(5)

			if pingErr := opened.Ping(); pingErr != nil {
				if cfg.RequirePostgres {
					panic(fmt.Sprintf("postgres unavailable: %v", pingErr))
				}
				log.Printf("postgres unavailable, fallback to in-memory services: %v", pingErr)
				_ = opened.Close()
			} else {
				db = opened
				log.Printf("postgres connected")
			}
		}
	} else if cfg.RequirePostgres {
		panic("postgres is required but DATABASE_URL is empty")
	}

	if db != nil {
		if err := events.EnsureSchema(db); err != nil {
			if cfg.RequirePostgres {
				panic(fmt.Sprintf("failed to ensure events schema: %v", err))
			}
			log.Printf("failed to ensure events schema: %v", err)
		}
	}

	authService := auth.NewServiceWithDB(db)
	if cfg.BootstrapAdminEmail != "" && cfg.BootstrapAdminPassword != "" {
		adminUser, err := authService.BootstrapAdmin(cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword)
		if err != nil {
			log.Printf("failed to bootstrap admin user: %v", err)
		} else {
			log.Printf("bootstrap admin ready: email=%s role=%s", adminUser.Email, adminUser.Role)
		}
	}

	eventsService := events.NewServiceWithDB(db)
	walletService := wallet.NewServiceWithDB(db, 1000)
	betsService := bets.NewServiceWithDB(db, eventsService, walletService)
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
			eventsGroup.GET("/:id/odds", eventsHandler.GetEventOdds)

			eventsAuthGroup := eventsGroup.Group("")
			eventsAuthGroup.Use(middleware.RequireAuth(cfg.AuthJWTSecret))
			{
				eventsAuthGroup.POST("", eventsHandler.CreateEvent)
				eventsAuthGroup.POST("/:id/request-settlement", eventsHandler.RequestSettlement)
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
			adminGroup.GET("/events/settlement-requests", eventsHandler.ListSettlementRequests)
			adminGroup.POST("/events/:id/settle", eventsHandler.SettleEvent)
		}
	}

	return router
}
