package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"walkover/server/config"
	"walkover/server/internal/extraction"
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
	extractor, _ := extraction.NewExtractor(context.Background(), "", "", "")
	handler.WithExtractor(extractor)
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

func TestTwilioFormWebhookNormalizesAndIngestsCivicIssue(t *testing.T) {
	handler, _, router := setupTestRouter()

	form := url.Values{
		"MessageSid":  {"SM-test-001"},
		"From":        {"whatsapp:+916232230297"},
		"To":          {"whatsapp:+14155238886"},
		"Body":        {"Vijay Nagar road par bada pothole hai aur ambulance ko problem ho rahi hai"},
		"ProfileName": {"Nidhi Agrawal"},
		"WaId":        {"916232230297"},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/viasocket", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()

	handler.mutex.RLock()
	before := len(handler.signals)
	handler.mutex.RUnlock()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	handler.mutex.RLock()
	defer handler.mutex.RUnlock()
	if len(handler.signals) != before+1 {
		t.Fatalf("expected one verified signal, before=%d after=%d", before, len(handler.signals))
	}
	if handler.signals[0].SenderPhone != "+916232230297" {
		t.Fatalf("expected normalized sender, got %q", handler.signals[0].SenderPhone)
	}
}

func TestWebhookRejectsGreetingFromOperationalFeed(t *testing.T) {
	handler, _, router := setupTestRouter()
	handler.mutex.RLock()
	before := len(handler.signals)
	handler.mutex.RUnlock()

	body := `{"provider":"WhatsApp","sender":"whatsapp:+916232230297","body":"Hi"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/viasocket", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 acknowledgement, got %d: %s", resp.Code, resp.Body.String())
	}
	handler.mutex.RLock()
	after := len(handler.signals)
	handler.mutex.RUnlock()
	if after != before {
		t.Fatalf("greeting entered operational feed: before=%d after=%d", before, after)
	}
}

func TestFireDoesNotMergeIntoManholeCluster(t *testing.T) {
	cfg := &config.Config{Port: "8080", Env: "development", AllowedOrigins: []string{"*"}}
	hub := NewHub()
	go hub.Run()
	handler := NewHandler(cfg, hub)
	extractor, _ := extraction.NewExtractor(context.Background(), "", "", "")
	handler.WithExtractor(extractor)
	now := time.Now()
	handler.ingest(models.CitizenSignal{ID: "sig-live-manhole", Provider: models.ProviderTelegram, RawText: "Khajrane ke yaha manhole khula hai", Timestamp: now}, nil)
	handler.ingest(models.CitizenSignal{ID: "sig-live-fire", Provider: models.ProviderTelegram, RawText: "Khajrana mandir ke paas aag lagi hai", Timestamp: now.Add(time.Second)}, nil)
	handler.mutex.RLock()
	defer handler.mutex.RUnlock()
	if len(handler.clusters) != 2 {
		t.Fatalf("fire and manhole merged; got %d clusters", len(handler.clusters))
	}
}

func TestFeedReturnsEnrichedPrivacySafeView(t *testing.T) {
	_, _, router := setupTestRouter()
	body := `{"provider":"WhatsApp","sender":"whatsapp:+911234567890","body":"Khajrana me open manhole hai"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/viasocket", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), request)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/feed", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"issue_category"`) || !strings.Contains(response.Body.String(), `"ward_name"`) {
		t.Fatalf("feed is not enriched: %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "+911234567890") || strings.Contains(response.Body.String(), "sender_phone") {
		t.Fatalf("feed exposed sender identity: %s", response.Body.String())
	}
}

func TestFirstVerifiedIssueCreatesPendingCluster(t *testing.T) {
	cfg := &config.Config{Port: "8080", Env: "development", AllowedOrigins: []string{"*"}}
	hub := NewHub()
	go hub.Run()
	handler := NewHandler(cfg, hub)
	extractor, _ := extraction.NewExtractor(context.Background(), "", "", "")
	handler.WithExtractor(extractor)
	router := SetupRouter(cfg, handler, hub)

	body := `{"provider":"WhatsApp","sender":"+919876543210","body":"Khajrana school gate ke paas open manhole hai"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/viasocket", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	handler.mutex.RLock()
	defer handler.mutex.RUnlock()
	if len(handler.clusters) != 1 {
		t.Fatalf("expected first verified issue to create one cluster, got %d", len(handler.clusters))
	}
	if handler.clusters[0].Status != "PENDING" || handler.clusters[0].SignalCount != 1 {
		t.Fatalf("unexpected new cluster: %+v", handler.clusters[0])
	}
}

func TestTelegramWebhookIngestsVerifiedIssue(t *testing.T) {
	handler, _, router := setupTestRouter()
	handler.mutex.RLock()
	before := len(handler.signals)
	handler.mutex.RUnlock()

	body := `{
		"update_id": 9001,
		"message": {
			"message_id": 51,
			"date": 1788681600,
			"text": "Khajrana school gate ke paas open manhole hai",
			"from": {"id": 123456, "first_name": "Nidhi", "username": "civic_test"},
			"chat": {"id": 123456, "type": "private"}
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/telegram", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	handler.mutex.RLock()
	defer handler.mutex.RUnlock()
	if len(handler.signals) != before+1 {
		t.Fatalf("expected Telegram issue in verified feed, before=%d after=%d", before, len(handler.signals))
	}
	if handler.signals[0].Provider != models.ProviderTelegram {
		t.Fatalf("expected Telegram provider, got %q", handler.signals[0].Provider)
	}
}

func TestTelegramGreetingDoesNotEnterOperationalFeed(t *testing.T) {
	handler, _, router := setupTestRouter()
	handler.mutex.RLock()
	before := len(handler.signals)
	handler.mutex.RUnlock()

	body := `{"update_id":9002,"message":{"message_id":52,"text":"Hello","from":{"id":123456},"chat":{"id":123456,"type":"private"}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/telegram", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 acknowledgement, got %d", resp.Code)
	}
	handler.mutex.RLock()
	after := len(handler.signals)
	handler.mutex.RUnlock()
	if after != before {
		t.Fatalf("Telegram greeting entered operational feed: before=%d after=%d", before, after)
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
