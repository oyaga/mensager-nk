package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nakamura/chatwoot-go/internal/config"
	"github.com/nakamura/chatwoot-go/internal/handlers"
	"github.com/nakamura/chatwoot-go/internal/middleware"
	"github.com/nakamura/chatwoot-go/internal/storage"
	"github.com/nakamura/chatwoot-go/internal/websocket"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	router *gin.Engine,
	db *gorm.DB,
	redis *redis.Client,
	wsHub *websocket.Hub,
	storageService *storage.MinioService,
	cfg *config.Config,
) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	accountHandler := handlers.NewAccountHandler(db)
	conversationHandler := handlers.NewConversationHandler(db, wsHub)
	messageHandler := handlers.NewMessageHandler(db, wsHub)
	contactHandler := handlers.NewContactHandler(db)
	inboxHandler := handlers.NewInboxHandler(db)
	uploadHandler := handlers.NewUploadHandler(storageService)
	wsHandler := handlers.NewWebSocketHandler(wsHub, cfg)
	webhookHandler := handlers.NewWebhookHandler(db)
	incomingWebhookHandler := handlers.NewIncomingWebhookHandler(db, wsHub)

	// Public routes
	public := router.Group("/api/v1")
	{
		// Authentication
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/forgot-password", authHandler.ForgotPassword)
		public.POST("/auth/reset-password", authHandler.ResetPassword)

		// Public API (for widget)
		public.POST("/widget/contacts", contactHandler.CreatePublicContact)
		public.POST("/widget/conversations", conversationHandler.CreatePublicConversation)
		public.POST("/widget/messages", messageHandler.CreatePublicMessage)

		// Incoming webhooks (authenticated via API token in header)
		// Using wildcard to support both /:instance and /:account_id/:instance patterns without Gin conflicts
		public.POST("/webhooks/incoming/*pathParam", incomingWebhookHandler.HandleIncoming)
	}

	// Protected routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(cfg))
	{
		// Storage
		api.POST("/storage/upload", uploadHandler.Upload)

		// Profile
		api.GET("/profile", authHandler.GetProfile)
		api.PUT("/profile", authHandler.UpdateProfile)
		api.PUT("/profile/password", authHandler.ChangePassword)
		api.PUT("/profile/availability", authHandler.UpdateAvailability)
		api.POST("/profile/access_token", authHandler.ResetAccessToken)

		// API Tokens (Legacy removed, using User AccessToken)

		// Accounts
		accounts := api.Group("/accounts")
		{
			accounts.GET("", accountHandler.List)
			accounts.POST("", accountHandler.Create)
			accounts.GET("/:id", accountHandler.Get)
			accounts.PUT("/:id", accountHandler.Update)
			accounts.DELETE("/:id", accountHandler.Delete)

			// Account users
			accounts.GET("/:id/users", accountHandler.ListUsers)
			accounts.POST("/:id/users", accountHandler.AddUser)
			accounts.DELETE("/:id/users/:user_id", accountHandler.RemoveUser)

			// Account webhooks
			accounts.GET("/:id/webhooks", webhookHandler.List)
			accounts.POST("/:id/webhooks", webhookHandler.Create)
			accounts.PUT("/:id/webhooks/:webhook_id", webhookHandler.Update)
			accounts.DELETE("/:id/webhooks/:webhook_id", webhookHandler.Delete)
		}

		// Conversations
		conversations := api.Group("/conversations")
		{
			conversations.GET("", conversationHandler.List)
			conversations.POST("", conversationHandler.Create)
			conversations.GET("/:id", conversationHandler.Get)
			conversations.PUT("/:id", conversationHandler.Update)
			conversations.DELETE("/:id", conversationHandler.Delete)
			conversations.POST("/:id/assign", conversationHandler.Assign)
			conversations.POST("/:id/resolve", conversationHandler.Resolve)
			conversations.POST("/:id/reopen", conversationHandler.Reopen)
			conversations.POST("/:id/snooze", conversationHandler.Snooze)
			conversations.POST("/:id/labels", conversationHandler.AddLabel)
			conversations.DELETE("/:id/labels/:label_id", conversationHandler.RemoveLabel)
			conversations.GET("/:id/messages", messageHandler.ListByConversation)
		}

		// Messages
		messages := api.Group("/messages")
		{
			messages.POST("", messageHandler.Create)
			messages.GET("/:id", messageHandler.Get)
			messages.PUT("/:id", messageHandler.Update)
			messages.DELETE("/:id", messageHandler.Delete)
		}

		// Contacts
		contacts := api.Group("/contacts")
		{
			contacts.GET("", contactHandler.List)
			contacts.POST("", contactHandler.Create)
			contacts.GET("/:id", contactHandler.Get)
			contacts.PUT("/:id", contactHandler.Update)
			contacts.DELETE("/:id", contactHandler.Delete)
			contacts.GET("/:id/conversations", conversationHandler.ListByContact)
		}

		// Inboxes
		inboxes := api.Group("/inboxes")
		{
			inboxes.GET("", inboxHandler.List)
			inboxes.POST("", inboxHandler.Create)
			inboxes.GET("/:id", inboxHandler.Get)
			inboxes.PUT("/:id", inboxHandler.Update)
			inboxes.DELETE("/:id", inboxHandler.Delete)
		}

		// Webhooks (at account level)
		webhooks := api.Group("/webhooks")
		{
			webhooks.GET("", webhookHandler.List)
			webhooks.POST("", webhookHandler.Create)
			webhooks.PUT("/:id", webhookHandler.Update)
			webhooks.DELETE("/:id", webhookHandler.Delete)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.RequireRole("administrator"))
		{
			admin.GET("/stats", accountHandler.GetStats)
			admin.GET("/users", authHandler.ListUsers)
			admin.PUT("/users/:id/role", authHandler.UpdateUserRole)
		}
	}

	// WebSocket endpoint
	router.GET("/cable", wsHandler.HandleWebSocket)
}
