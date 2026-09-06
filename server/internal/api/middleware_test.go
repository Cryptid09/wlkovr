package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"walkover/server/config"
)

const testSecret = "test-webhook-secret"

func setupSecuredRouter(secret string) http.Handler {
	cfg := &config.Config{
		Port:                   "8080",
		Env:                    "test",
		AllowedOrigins:         []string{"*"},
		ViasocketWebhookSecret: secret,
	}
	hub := NewHub()
	go hub.Run()
	return SetupRouter(cfg, NewHandler(cfg, hub), hub)
}

func postWebhook(router http.Handler, header string) *httptest.ResponseRecorder {
	body := []byte(`{"provider":"WhatsApp","sender":"+91-90000-00001","body":"test complaint"}`)
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/viasocket", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if header != "" {
		req.Header.Set(WebhookSecretHeader, header)
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

// The ingestion endpoint is publicly reachable once tunnelled, so an
// unauthenticated caller must not be able to inject citizen signals.
func TestWebhookRejectsMissingSecret(t *testing.T) {
	if got := postWebhook(setupSecuredRouter(testSecret), "").Code; got != http.StatusUnauthorized {
		t.Errorf("webhook without a secret returned %d, want %d", got, http.StatusUnauthorized)
	}
}

func TestWebhookRejectsWrongSecret(t *testing.T) {
	if got := postWebhook(setupSecuredRouter(testSecret), "not-the-secret").Code; got != http.StatusUnauthorized {
		t.Errorf("webhook with a wrong secret returned %d, want %d", got, http.StatusUnauthorized)
	}
}

func TestWebhookAcceptsCorrectSecret(t *testing.T) {
	if got := postWebhook(setupSecuredRouter(testSecret), testSecret).Code; got != http.StatusOK {
		t.Errorf("webhook with the correct secret returned %d, want %d", got, http.StatusOK)
	}
}

// An unset secret leaves local development unauthenticated on purpose.
func TestWebhookOpenWhenNoSecretConfigured(t *testing.T) {
	if got := postWebhook(setupSecuredRouter(""), "").Code; got != http.StatusOK {
		t.Errorf("webhook with no configured secret returned %d, want %d", got, http.StatusOK)
	}
}

// Only the webhook is gated — the dashboard's own endpoints must stay open.
func TestSecretDoesNotGateDashboardEndpoints(t *testing.T) {
	router := setupSecuredRouter(testSecret)
	for _, path := range []string{"/api/v1/health", "/api/v1/wards", "/api/v1/clusters", "/api/v1/signals"} {
		req, _ := http.NewRequest("GET", path, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Errorf("GET %s returned %d, want %d", path, resp.Code, http.StatusOK)
		}
	}
}
