package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all server configuration parameters
type Config struct {
	Port                   string
	Env                    string
	AllowedOrigins         []string
	GeminiAPIKey           string
	GeminiModel            string
	EmbeddingModel         string
	GoogleCloudProject     string
	FirestoreEmulatorHost  string
	ViasocketWebhookSecret string
	TelegramBotToken       string
	TelegramWebhookSecret  string
	TwilioAccountSID       string
	TwilioAuthToken        string
	TwilioWhatsAppFrom     string
}

// LoadConfig loads configuration from .env and environment variables
func LoadConfig() *Config {
	// Try loading from .env in server directory or root directory
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("[CONFIG] No .env file found, using system environment variables")
		}
	}

	origins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")
	allowedOrigins := strings.Split(origins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		AllowedOrigins: allowedOrigins,
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		// gemini-2.5-flash returned "no longer available to new users" mid-build;
		// Google's own error names gemini-3.6-flash as the replacement.
		GeminiModel: getEnv("GEMINI_MODEL", "gemini-3.6-flash"),
		// text-embedding-004 is not served on the v1beta generativelanguage
		// endpoint this SDK uses — it 404s. gemini-embedding-001 is the model
		// that actually responds (3072 dimensions).
		EmbeddingModel:         getEnv("EMBEDDING_MODEL", "gemini-embedding-001"),
		GoogleCloudProject:     getEnv("GOOGLE_CLOUD_PROJECT", "zen-dev-intelligence"),
		FirestoreEmulatorHost:  getEnv("FIRESTORE_EMULATOR_HOST", ""),
		ViasocketWebhookSecret: getEnv("VIASOCKET_WEBHOOK_SECRET", "zen-secret-key-12345"),
		TelegramBotToken:       getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWebhookSecret:  getEnv("TELEGRAM_WEBHOOK_SECRET", ""),
		TwilioAccountSID:       getEnv("TWILIO_ACCOUNT_SID", ""),
		// No default: without an auth token the platform simply does not send
		// WhatsApp replies. A placeholder would look configured and fail late.
		TwilioAuthToken:    getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioWhatsAppFrom: getEnv("TWILIO_WHATSAPP_FROM", "whatsapp:+14155238886"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
