package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"tech-nova/internal/ai"
	"tech-nova/internal/chat"
	"tech-nova/internal/config"
	"tech-nova/internal/database"
	"tech-nova/internal/telegram"
	"tech-nova/internal/tenant"
	"tech-nova/internal/ticket"
	"tech-nova/internal/websocket"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. Initialize Database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization error: %v", err)
	}
	defer db.Close()

	// 2. Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 3. Initialize Repositories
	tenantRepo := tenant.NewRepository(db)
	ticketRepo := ticket.NewRepository(db)
	chatRepo := chat.NewRepository(db)

	// 4. Initialize Services
	aiService := ai.NewService(cfg)
	tgService := telegram.NewService(cfg)
	tenantService := tenant.NewService(tenantRepo)
	ticketService := ticket.NewService(ticketRepo)
	chatService := chat.NewService(cfg, chatRepo, ticketRepo, aiService, tgService, wsHub)

	// 5. Initialize Handlers
	ticketHandler := ticket.NewHandler(ticketService, tenantService)
	chatHandler := chat.NewHandler(chatService)
	tgHandler := telegram.NewHandler(tgService, ticketRepo, chatService, wsHub)

	// 6. Router Setup
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Key-Identifier")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Static files (Uploads)
	r.Static("/uploads", cfg.UploadDir)

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": cfg.AppName,
			"env":     cfg.AppEnv,
		})
	})

	// WebSocket Endpoint
	r.GET("/ws", websocket.ServeWS(wsHub))

	// Telegram Webhook
	r.POST("/webhook/telegram", tgHandler.HandleWebhook)

	// REST API v1
	v1 := r.Group("/api/v1")
	{
		// Tickets
		v1.POST("/tickets/init", ticketHandler.InitTicket)
		v1.POST("/tickets/:ticket_id/rate", ticketHandler.RateTicket)

		// Chat & Attachments
		v1.POST("/chat/send", chatHandler.SendMessage)
		v1.POST("/chat/upload", chatHandler.UploadAttachment)
		v1.GET("/tickets/:ticket_id/messages", chatHandler.GetHistory)
	}

	log.Printf("🚀 [%s] Universal Helpdesk & AI Gateway running on http://localhost:%s", cfg.AppName, cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
