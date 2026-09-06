package api

import (
	"crypto/subtle"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"walkover/server/config"
	"walkover/server/internal/models"
)

const TelegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token"

func RequireTelegramSecret(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.TelegramWebhookSecret == "" || subtle.ConstantTimeCompare([]byte(c.GetHeader(TelegramSecretHeader)), []byte(cfg.TelegramWebhookSecret)) == 1 {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ApiResponse{Success: false, Error: "Unauthorized Telegram webhook"})
	}
}

// WebhookSecretHeader is the header viasocket must send on every forwarded
// citizen signal. Configure it as a custom header on the flow's HTTP action.
const WebhookSecretHeader = "X-Webhook-Secret"

// RequireWebhookSecret rejects webhook calls that do not carry the shared
// secret.
//
// The ingestion endpoint has to be publicly reachable for viasocket to deliver
// to it, which also means anyone who discovers the URL could inject fabricated
// citizen signals into the platform. A shared secret is the minimum bar; it is
// not a substitute for real request signing.
//
// An empty configured secret disables the check, so local development without
// a secret keeps working.
func RequireWebhookSecret(cfg *config.Config) gin.HandlerFunc {
	expected := cfg.ViasocketWebhookSecret

	return func(c *gin.Context) {
		if expected == "" {
			c.Next()
			return
		}

		provided := c.GetHeader(WebhookSecretHeader)
		// Constant-time comparison so a caller cannot recover the secret by
		// measuring how long a rejection takes.
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			log.Printf("[SECURITY] Rejected webhook from %s: missing or invalid %s", c.ClientIP(), WebhookSecretHeader)
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ApiResponse{
				Success: false,
				Error:   "Unauthorized: missing or invalid " + WebhookSecretHeader,
			})
			return
		}

		c.Next()
	}
}
