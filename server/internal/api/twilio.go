package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"walkover/server/internal/models"
)

// Twilio delivers WhatsApp messages as form-encoded fields, not as the JSON
// shape viasocket normalises to. Accepting both means the same endpoint works
// whether the viasocket flow transforms the payload or forwards it untouched,
// which removes a whole class of "the flow is configured slightly wrong"
// failures on demo day.

const twilioSendTimeout = 15 * time.Second

// twilioNumberPrefix is how Twilio addresses WhatsApp endpoints.
const twilioNumberPrefix = "whatsapp:"

// isTwilioForm reports whether the request carries a Twilio webhook rather than
// a viasocket-normalised JSON body.
func isTwilioForm(c *gin.Context) bool {
	contentType := c.ContentType()
	if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data") {
		// Only treat it as Twilio if it actually looks like one.
		return c.PostForm("Body") != "" || c.PostForm("From") != "" || c.PostForm("MessageSid") != ""
	}
	return false
}

// payloadFromTwilio maps a Twilio WhatsApp/SMS webhook onto the platform's
// canonical payload.
func payloadFromTwilio(c *gin.Context) models.ViasocketPayload {
	from := c.PostForm("From")

	// A "whatsapp:" prefix is how Twilio distinguishes WhatsApp from SMS on the
	// same account.
	provider := "SMS"
	if strings.HasPrefix(from, twilioNumberPrefix) {
		provider = "WhatsApp"
	}

	metadata := map[string]interface{}{
		"channel": "twilio",
		"to":      c.PostForm("To"),
	}
	for field, key := range map[string]string{
		"ProfileName": "profile_name",
		"WaId":        "wa_id",
		"MessageSid":  "message_sid",
		"MessageType": "message_type",
		"AccountSid":  "account_sid",
	} {
		if value := c.PostForm(field); value != "" {
			metadata[key] = value
		}
	}

	// Voice notes and photos arrive as media rather than text. The URL is kept
	// so evidence handling can pick it up later; transcription is not built.
	mediaURL := c.PostForm("MediaUrl0")
	if mediaURL != "" {
		metadata["media_content_type"] = c.PostForm("MediaContentType0")
	}

	return models.ViasocketPayload{
		EventID:   c.PostForm("MessageSid"),
		Provider:  provider,
		Sender:    strings.TrimPrefix(from, twilioNumberPrefix),
		Body:      c.PostForm("Body"),
		MediaURL:  mediaURL,
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata:  metadata,
	}
}

// replyToCitizen sends an acknowledgement back over WhatsApp, closing the loop
// the pitch promises: a citizen who reports something hears that it was
// understood and where it was routed.
//
// It is deliberately an acknowledgement, not a decision. It never tells a
// citizen that work is approved or scheduled — only that their report was
// received, understood, and grouped with corroborating reports.
func (h *Handler) replyToCitizen(signal models.CitizenSignal, extraction *models.AIExtraction, corroborating int) {
	if signal.Provider != models.ProviderWhatsApp {
		return
	}
	if h.cfg.TwilioAccountSID == "" || h.cfg.TwilioAuthToken == "" {
		return
	}
	if signal.SenderPhone == "" {
		return
	}

	message := composeAcknowledgement(extraction, corroborating)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), twilioSendTimeout)
		defer cancel()

		if err := h.sendTwilioMessage(ctx, signal.SenderPhone, message); err != nil {
			log.Printf("[TWILIO] Could not reply to %s: %v", signal.SenderPhone, err)
			return
		}
		log.Printf("[TWILIO] Acknowledged %s to %s", signal.ID, signal.SenderPhone)
	}()
}

// composeAcknowledgement writes what the platform actually understood, so the
// citizen can tell whether they were misread.
func composeAcknowledgement(extraction *models.AIExtraction, corroborating int) string {
	if extraction == nil || extraction.Issue == "" {
		return "Thank you — your report has been received and recorded for review by the municipal team. No action has been approved yet; an official will review the evidence."
	}

	var b strings.Builder
	b.WriteString("Thank you. Your report has been recorded.\n\n")
	fmt.Fprintf(&b, "Understood as: %s\n", extraction.Issue)
	if extraction.WardName != "" {
		fmt.Fprintf(&b, "Location: %s\n", extraction.WardName)
	}
	if extraction.Department != "" {
		fmt.Fprintf(&b, "Routed to: %s\n", extraction.Department)
	}
	if corroborating > 1 {
		fmt.Fprintf(&b, "\nThis matches %d other reports from your area, which strengthens the case for investigation.", corroborating)
	}
	b.WriteString("\n\nAn official will review the evidence. Nothing has been approved automatically. If we misunderstood your report, please reply with more detail.")
	return b.String()
}

// sendTwilioMessage posts to the Twilio Messages API.
func (h *Handler) sendTwilioMessage(ctx context.Context, to, body string) error {
	if !strings.HasPrefix(to, twilioNumberPrefix) {
		to = twilioNumberPrefix + to
	}

	form := url.Values{}
	form.Set("From", h.cfg.TwilioWhatsAppFrom)
	form.Set("To", to)
	form.Set("Body", body)

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", h.cfg.TwilioAccountSID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(h.cfg.TwilioAccountSID, h.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		// Twilio puts the actionable reason in the body (its own error code and
		// message). A bare status line is not enough to diagnose a failure
		// during a demo.
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		body := strings.TrimSpace(string(detail))

		// 21654 / 63016 mean WhatsApp's 24-hour session window is closed: a
		// business may only send free-form text within 24 hours of the citizen's
		// own message, and must use an approved template outside it. This is the
		// expected result of replaying a synthetic webhook, since no real
		// inbound WhatsApp message opened a window.
		if strings.Contains(body, "21654") || strings.Contains(body, "63016") {
			return fmt.Errorf("no open WhatsApp session with %s — the citizen must message the sandbox first, or an approved template (TWILIO_CONTENT_SID) is required outside the 24-hour window (twilio: %s)", to, body)
		}
		return fmt.Errorf("twilio responded %s: %s", resp.Status, body)
	}
	return nil
}
