package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"walkover/server/config"
)

// SetupRouter configures all Gin routes, middlewares, and WebSocket endpoints
func SetupRouter(cfg *config.Config, handler *Handler, hub *Hub) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Configure CORS for Next.js frontend
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// WebSocket endpoint
	r.GET("/ws", hub.HandleWebSocket)

	// API v1 Routes
	v1 := r.Group("/api/v1")
	{
		// Health
		v1.GET("/health", handler.HealthCheck)

		// Wards & Maps
		v1.GET("/wards", handler.GetWards)

		// Issue Clusters & Priority Hotspots
		v1.GET("/clusters", handler.GetClusters)
		v1.GET("/clusters/:id", handler.GetClusterByID)

		// Citizen Signals Live Feed
		v1.GET("/signals", handler.GetSignals)

		// Webhooks (viasocket WhatsApp/SMS ingestion) — publicly reachable, so
		// it is gated by the shared secret viasocket sends as a custom header.
		v1.POST("/webhooks/viasocket", RequireWebhookSecret(cfg), handler.HandleViasocketWebhook)
		v1.POST("/webhooks/telegram", RequireTelegramSecret(cfg), handler.HandleTelegramWebhook)

		// Policymaker Human Decision Action Panel
		v1.POST("/decisions", handler.RecordDecision)
		v1.GET("/audit-logs", handler.GetAuditLogs)

		// Policymaker assistant — grounded Q&A over the evidence on screen
		v1.POST("/assistant", handler.AskAssistant)

		// Live Simulation for Demo testing
		v1.POST("/demo/simulate", handler.SimulateLiveMessage)
	}

	return r
}
