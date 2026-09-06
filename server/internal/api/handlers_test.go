package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"walkover/server/config"
	"walkover/server/internal/models"
)

func setupTestRouter() (*Handler, *Hub, http.Handler) {
	cfg := &config.Config{
		Port:           "8080",
		Env:            "test",
		AllowedOrigins: []string{"*"},
	}
	hub := NewHub()
	go hub.Run()

	handler := NewHandler(cfg, hub)
	router := SetupRouter(cfg, handler, hub)
	return handler, hub, router
}

func TestHealthCheck(t *testing.T) {
	_, _, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var apiResp models.ApiResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !apiResp.Success {
		t.Fatalf("expected success true, got false")
	}
}

func TestHandleViasocketWebhook_WhatsApp(t *testing.T) {
	handler, _, router := setupTestRouter()

	payload := models.ViasocketPayload{
		EventID:   "test-evt-001",
		Provider:  "WhatsApp",
		Sender:    "+91-9826012345",
		Body:      "हमारे यहाँ नल से गंदा पानी आ रहा है बहुत बदबू है चंदन नगर गली 4",
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"sender_name":   "Ramesh Sharma",
			"location_hint": "Chandan Nagar Street 4",
			"ward_id":       "indore-ward-02",
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/viasocket", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	elapsed := time.Since(start)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}

	var apiResp models.ApiResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if !apiResp.Success {
		t.Fatalf("expected success true in apiResp")
	}

	// Verify signal was saved in memory
	handler.mutex.RLock()
	signalsCount := len(handler.signals)
	latestSignal := handler.signals[0]
	handler.mutex.RUnlock()

	if signalsCount == 0 {
		t.Fatalf("expected signals count > 0, got 0")
	}
	if latestSignal.Provider != models.ProviderWhatsApp {
		t.Errorf("expected provider WhatsApp, got %v", latestSignal.Provider)
	}
	if latestSignal.RawText != payload.Body {
		t.Errorf("expected raw text match, got %v", latestSignal.RawText)
	}
	if latestSignal.SenderPhone != payload.Sender {
		t.Errorf("expected sender phone %v, got %v", payload.Sender, latestSignal.SenderPhone)
	}

	// Latency requirement: Ingestion round-trip must be < 2 seconds (< 50ms typical)
	if elapsed > 2*time.Second {
		t.Errorf("expected latency < 2s, took %v", elapsed)
	}
}

func TestHandleViasocketWebhook_SMS(t *testing.T) {
	handler, _, router := setupTestRouter()

	payload := models.ViasocketPayload{
		EventID:   "test-evt-sms-001",
		Provider:  "SMS",
		Sender:    "+91-9425098765",
		Body:      "Vijay Nagar square ke paas road dhas gayi hai",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/viasocket", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	handler.mutex.RLock()
	latestSignal := handler.signals[0]
	handler.mutex.RUnlock()

	if latestSignal.Provider != models.ProviderSMS {
		t.Errorf("expected provider SMS, got %v", latestSignal.Provider)
	}
}

func TestHandleViasocketWebhook_InvalidPayload(t *testing.T) {
	_, _, router := setupTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/webhooks/viasocket", bytes.NewBuffer([]byte(`{malformed_json`)))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for malformed json, got %d", resp.Code)
	}
}

func TestGetSignals(t *testing.T) {
	_, _, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/signals", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var apiResp models.ApiResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &apiResp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if !apiResp.Success {
		t.Fatalf("expected success true")
	}
}

func TestGetClustersAndByID(t *testing.T) {
	_, _, router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/clusters", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	// Test Get by ID
	req2, _ := http.NewRequest("GET", "/api/v1/clusters/cluster-indore-001", nil)
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)

	if resp2.Code != http.StatusOK {
		t.Fatalf("expected status 200 for cluster-indore-001, got %d", resp2.Code)
	}

	// Test Not Found
	req3, _ := http.NewRequest("GET", "/api/v1/clusters/non-existent-id", nil)
	resp3 := httptest.NewRecorder()
	router.ServeHTTP(resp3, req3)

	if resp3.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for non-existent cluster, got %d", resp3.Code)
	}
}

func TestRecordDecision(t *testing.T) {
	_, _, router := setupTestRouter()

	decisionReq := models.DecisionRequest{
		ClusterID: "cluster-indore-001",
		Action:    "ACCEPTED",
		Officer:   "Commissioner Municipal Corp",
		Notes:     "Emergency drainage team dispatched with vacuum suction units.",
	}

	bodyBytes, _ := json.Marshal(decisionReq)
	req, _ := http.NewRequest("POST", "/api/v1/decisions", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}

	// Check audit log
	reqLog, _ := http.NewRequest("GET", "/api/v1/audit-logs", nil)
	respLog := httptest.NewRecorder()
	router.ServeHTTP(respLog, reqLog)

	if respLog.Code != http.StatusOK {
		t.Fatalf("expected status 200 for audit logs, got %d", respLog.Code)
	}
}

func TestSimulateLiveMessage(t *testing.T) {
	_, _, router := setupTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/demo/simulate", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200 for simulate live message, got %d", resp.Code)
	}
}
