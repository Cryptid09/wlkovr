package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"walkover/server/config"
	"walkover/server/internal/models"
)

// The exact field set Twilio sends for a WhatsApp sandbox message, taken from
// the verified payload in twilio_viasocket_whatsapp_setup.md.
func twilioForm(body, from string) url.Values {
	return url.Values{
		"Body":        {body},
		"From":        {from},
		"To":          {"whatsapp:+14155238886"},
		"SmsStatus":   {"received"},
		"MessageType": {"text"},
		"ProfileName": {"Nidhi Agrawal"},
		"WaId":        {"916232230297"},
		"NumMedia":    {"0"},
		"MessageSid":  {"SM1234567890abcdef"},
		"AccountSid":  {"ACtest00000000000000000000000000000"},
	}
}

func postForm(router http.Handler, form url.Values) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/viasocket", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func TestWebhookAcceptsTwilioForm(t *testing.T) {
	handler, _, router := setupTestRouter()

	before := len(handler.signals)
	resp := postForm(router, twilioForm("Khajrana me manhole khula hai", "whatsapp:+916232230297"))
	if resp.Code != http.StatusOK {
		t.Fatalf("Twilio form webhook returned %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	handler.mutex.RLock()
	defer handler.mutex.RUnlock()
	if len(handler.signals) != before+1 {
		t.Fatalf("signal count went %d -> %d, want one new signal", before, len(handler.signals))
	}

	signal := handler.signals[0]
	if signal.RawText != "Khajrana me manhole khula hai" {
		t.Errorf("RawText = %q", signal.RawText)
	}
	// The whatsapp: prefix is Twilio addressing, not part of the phone number.
	if signal.SenderPhone != "+916232230297" {
		t.Errorf("SenderPhone = %q, want %q", signal.SenderPhone, "+916232230297")
	}
	if signal.Provider != models.ProviderWhatsApp {
		t.Errorf("Provider = %q, want %q", signal.Provider, models.ProviderWhatsApp)
	}
	if signal.Metadata["profile_name"] != "Nidhi Agrawal" {
		t.Errorf("profile_name metadata = %v", signal.Metadata["profile_name"])
	}
}

// A number without the whatsapp: prefix is an SMS on the same Twilio account.
func TestTwilioSMSDetectedAsSMS(t *testing.T) {
	handler, _, router := setupTestRouter()

	if code := postForm(router, twilioForm("Road cave-in near BRTS", "+916232230297")).Code; code != http.StatusOK {
		t.Fatalf("returned %d", code)
	}

	handler.mutex.RLock()
	defer handler.mutex.RUnlock()
	if handler.signals[0].Provider != models.ProviderSMS {
		t.Errorf("Provider = %q, want %q", handler.signals[0].Provider, models.ProviderSMS)
	}
}

// An empty Twilio body carries nothing to extract and must not become a signal.
func TestWebhookRejectsEmptyBody(t *testing.T) {
	_, _, router := setupTestRouter()

	form := twilioForm("", "whatsapp:+916232230297")
	if code := postForm(router, form).Code; code != http.StatusBadRequest {
		t.Errorf("empty body returned %d, want %d", code, http.StatusBadRequest)
	}
}

// The acknowledgement tells the citizen what was understood, and must never
// imply that work has been approved.
func TestAcknowledgementReportsUnderstandingNotApproval(t *testing.T) {
	message := composeAcknowledgement(&models.AIExtraction{
		Issue:      "Uncovered drainage manhole in pedestrian path",
		WardName:   "Ward 60 - Khajrana",
		Department: "Water Supply & Sewerage",
	}, 7)

	for _, want := range []string{"Uncovered drainage manhole", "Ward 60 - Khajrana", "Water Supply & Sewerage", "7 other reports"} {
		if !strings.Contains(message, want) {
			t.Errorf("acknowledgement missing %q: %s", want, message)
		}
	}
	const disclaimer = "Nothing has been approved automatically."
	if !strings.Contains(message, disclaimer) {
		t.Fatalf("acknowledgement should state that nothing was auto-approved: %s", message)
	}

	// Scan everything except the disclaimer, so approval language anywhere else
	// is caught even in its bare form.
	body := strings.ToLower(strings.ReplaceAll(message, disclaimer, ""))
	for _, promise := range []string{"approved", "sanctioned", "scheduled", "will be fixed", "will be repaired"} {
		if strings.Contains(body, promise) {
			t.Errorf("acknowledgement promises action (%q): %s", promise, message)
		}
	}
}

func TestAcknowledgementHandlesFailedExtraction(t *testing.T) {
	message := composeAcknowledgement(nil, 1)
	if message == "" {
		t.Fatal("expected an acknowledgement even without an extraction")
	}
	if strings.Contains(strings.ToLower(message), "approved yet") == false {
		t.Errorf("fallback acknowledgement should still disclaim approval: %s", message)
	}
}

// Without Twilio credentials the platform must simply not send, rather than
// erroring or blocking ingestion.
func TestReplySkippedWithoutCredentials(t *testing.T) {
	cfg := &config.Config{Env: "test", AllowedOrigins: []string{"*"}}
	hub := NewHub()
	go hub.Run()
	handler := NewHandler(cfg, hub)

	// Must not panic or block.
	handler.replyToCitizen(models.CitizenSignal{ID: "sig-1", SenderPhone: "+916232230297"}, nil, 1)
}
