package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yandas/backend/internal/config"
	"github.com/yandas/backend/internal/database"
	"github.com/yandas/backend/internal/handlers"
	"github.com/yandas/backend/internal/middleware"
	"github.com/yandas/backend/internal/repository"
	"github.com/yandas/backend/internal/services"
	"github.com/yandas/backend/internal/websocket"
)

// @title YANDAŞ API
// @version 1.0
// @description Yeni Nesil Vekalet ve Yerinden Hizmet Platformu API
// @termsOfService https://yandas.app/terms

// @contact.name API Support
// @contact.url https://yandas.app/support
// @contact.email support@yandas.app

// @license.name Proprietary
// @license.url https://yandas.app/license

// @host api.yandas.app
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	redisClient, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Printf("Failed to connect to Redis: %v (continuing without cache)", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	svcs := services.NewServices(repos, cfg, redisClient)

	// Seed initial data
	if err := database.Seed(db, cfg); err != nil {
		log.Printf("Failed to seed database: %v", err)
	}

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize handlers
	h := handlers.NewHandlers(svcs, cfg, wsHub, db)

	// Setup router
	router := gin.Default()

	// Set max multipart memory for file uploads (50 MB)
	router.MaxMultipartMemory = 50 << 20 // 50 MB

	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimiter(cfg, redisClient))
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": "1.0.0"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.RefreshToken)
			auth.POST("/forgot-password", h.Auth.ForgotPassword)
			auth.POST("/reset-password", h.Auth.ResetPassword)
			auth.POST("/verify-phone", h.Auth.VerifyPhone)
			auth.POST("/resend-otp", h.Auth.ResendOTP)
			auth.POST("/verify-account", h.Auth.VerifyAccount)
			auth.POST("/resend-email-otp", h.Auth.ResendEmailOTP)
		}

		// Categories (public)
		v1.GET("/categories", h.Category.List)

		// Public Yandaş listing
		v1.GET("/yandas", h.Yandas.ListPublic)
		v1.GET("/yandas/:id", h.Yandas.GetPublic)
		v1.GET("/yandas/:id/services", h.Yandas.GetServices)
		v1.GET("/yandas/:id/reviews", h.Yandas.GetReviews)

		// Search (public)
		v1.GET("/search", h.Search.SearchYandas)

		// Legal pages (public)
		legal := v1.Group("/legal")
		{
			legal.GET("/privacy", h.Legal.PrivacyPolicy)
			legal.GET("/terms", h.Legal.TermsOfService)
			legal.GET("/kvkk", h.Legal.KVKK)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(cfg))
		{
			// User profile
			user := protected.Group("/user")
			{
				user.GET("/me", h.User.GetProfile)
				user.PUT("/me", h.User.UpdateProfile)
				user.PUT("/me/avatar", h.User.UpdateAvatar)
				user.PUT("/me/password", h.User.ChangePassword)
				user.DELETE("/me", h.User.DeleteAccount)
				user.POST("/me/device-token", h.User.RegisterDeviceToken)
			}

			// Yandaş application & management
			yandas := protected.Group("/yandas")
			{
				yandas.POST("/apply", h.Yandas.Apply)
				yandas.GET("/application-status", h.Yandas.ApplicationStatus)
				yandas.PUT("/profile", h.Yandas.UpdateProfile)
				yandas.PUT("/availability", h.Yandas.UpdateAvailability)
				yandas.PUT("/location", h.Yandas.UpdateLocation)

				// Services management
				yandas.POST("/services", h.Yandas.CreateService)
				yandas.PUT("/services/:id", h.Yandas.UpdateService)
				yandas.DELETE("/services/:id", h.Yandas.DeleteService)
				yandas.GET("/my-services", h.Yandas.GetMyServices)

				// Incoming orders
				yandas.GET("/orders", h.Yandas.GetOrders)
				yandas.POST("/orders/:id/accept", h.Yandas.AcceptOrder)
				yandas.POST("/orders/:id/reject", h.Yandas.RejectOrder)
				yandas.POST("/orders/:id/start", h.Yandas.StartOrder)
				yandas.POST("/orders/:id/complete", h.Yandas.CompleteOrder)

				// Stats
				yandas.GET("/stats", h.Yandas.GetStats)
			}

			// Orders (customer side)
			orders := protected.Group("/orders")
			{
				orders.POST("", h.Order.Create)
				orders.GET("", h.Order.List)
				orders.GET("/:id", h.Order.Get)
				orders.POST("/:id/cancel", h.Order.Cancel)
				orders.POST("/:id/review", h.Order.Review)
			}

			// Chat
			chat := protected.Group("/chat")
			{
				chat.GET("/conversations", h.Chat.ListConversations)
				chat.POST("/conversations/start", h.Chat.StartConversation)
				chat.GET("/conversations/:id", h.Chat.GetConversation)
				chat.GET("/conversations/:id/messages", h.Chat.GetMessages)
				chat.POST("/conversations/:id/messages", h.Chat.SendMessage)
				chat.POST("/conversations/:id/read", h.Chat.MarkAsRead)
				chat.POST("/conversations/:id/image", h.Chat.SendImageMessage)
			}

			// Calls (voice/video)
			calls := protected.Group("/call")
			{
				calls.POST("/initiate", h.Call.InitiateCall)
				calls.POST("/:id/answer", h.Call.AnswerCall)
				calls.POST("/:id/reject", h.Call.RejectCall)
				calls.POST("/:id/end", h.Call.EndCall)
			}

			// Favorites
			favorites := protected.Group("/favorites")
			{
				favorites.GET("", h.Favorite.List)
				favorites.GET("/ids", h.Favorite.IDs)
				favorites.POST("/:id/toggle", h.Favorite.Toggle)
				favorites.GET("/:id/check", h.Favorite.Check)
			}

			// Support tickets (user-facing)
			support := protected.Group("/support")
			{
				support.POST("/tickets", h.Support.CreateTicket)
				support.GET("/tickets", h.Support.ListTickets)
				support.GET("/tickets/:id", h.Support.GetTicket)
				support.POST("/tickets/:id/reply", h.Support.ReplyTicket)
			}

			// Subscriptions
			subscription := protected.Group("/subscription")
			{
				subscription.GET("", h.Subscription.Get)
				subscription.POST("/verify", h.Subscription.Verify)
				subscription.POST("/webhook", h.Subscription.Webhook)
			}

			// Notifications
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", h.Notification.List)
				notifications.POST("/:id/read", h.Notification.MarkAsRead)
				notifications.POST("/read-all", h.Notification.MarkAllAsRead)
			}
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg))
		admin.Use(middleware.AdminRequired())
		{
			// Dashboard
			admin.GET("/dashboard", h.Admin.Dashboard)

			// User management
			admin.GET("/users", h.Admin.ListUsers)
			admin.GET("/users/:id", h.Admin.GetUser)
			admin.PUT("/users/:id", h.Admin.UpdateUser)
			admin.DELETE("/users/:id", h.Admin.DeleteUser)

			// Yandaş applications
			admin.GET("/applications", h.Admin.ListApplications)
			admin.GET("/applications/:id", h.Admin.GetApplication)
			admin.POST("/applications/:id/approve", h.Admin.ApproveApplication)
			admin.POST("/applications/:id/reject", h.Admin.RejectApplication)

			// Orders
			admin.GET("/orders", h.Admin.ListOrders)
			admin.GET("/orders/:id", h.Admin.GetOrder)

			// Categories
			admin.POST("/categories", h.Admin.CreateCategory)
			admin.PUT("/categories/:id", h.Admin.UpdateCategory)
			admin.DELETE("/categories/:id", h.Admin.DeleteCategory)

			// Analytics
			admin.GET("/analytics/overview", h.Admin.AnalyticsOverview)
			admin.GET("/analytics/revenue", h.Admin.AnalyticsRevenue)
			admin.GET("/analytics/users", h.Admin.AnalyticsUsers)

			// Audit logs
			admin.GET("/audit-logs", h.Admin.AuditLogs)

			// Support tickets
			admin.GET("/support/tickets", h.Admin.ListSupportTickets)
			admin.GET("/support/tickets/:id", h.Admin.GetSupportTicket)
			admin.PUT("/support/tickets/:id", h.Admin.UpdateSupportTicket)
			admin.POST("/support/tickets/:id/reply", h.Admin.ReplySupportTicket)
			admin.GET("/support/stats", h.Admin.GetSupportStats)
		}

		// WebSocket
		v1.GET("/ws", middleware.AuthRequired(cfg), func(c *gin.Context) {
			websocket.HandleConnection(wsHub, c)
		})
	}

	// Static files for uploads
	router.Static("/uploads", "./uploads")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 YANDAŞ API starting on 0.0.0.0:%s", port)
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
