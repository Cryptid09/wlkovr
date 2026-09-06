package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"walkover/server/config"
	"walkover/server/internal/models"
)

func setupAssistantRouter() (*Handler, http.Handler) {
	cfg := &config.Config{Port: "8080", Env: "test", AllowedOrigins: []string{"*"}}
	hub := NewHub()
	go hub.Run()
	handler := NewHandler(cfg, hub)
	return handler, SetupRouter(cfg, handler, hub)
}

func ask(t *testing.T, router http.Handler, body string) (int, AssistantResponse) {
	t.Helper()

	req, _ := http.NewRequest("POST", "/api/v1/assistant", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	var wrapper struct {
		Success bool              `json:"success"`
		Data    AssistantResponse `json:"data"`
	}
	json.Unmarshal(resp.Body.Bytes(), &wrapper)
	return resp.Code, wrapper.Data
}

func TestAssistantRequiresAQuestion(t *testing.T) {
	_, router := setupAssistantRouter()

	if code, _ := ask(t, router, `{"question":""}`); code != http.StatusBadRequest {
		t.Errorf("empty question returned %d, want %d", code, http.StatusBadRequest)
	}
	if code, _ := ask(t, router, `{}`); code != http.StatusBadRequest {
		t.Errorf("missing question returned %d, want %d", code, http.StatusBadRequest)
	}
}

// With no Gemini key configured the panel must still answer from stored
// evidence rather than returning an error or an empty bubble.
func TestAssistantAnswersOfflineFromStoredEvidence(t *testing.T) {
	handler, router := setupAssistantRouter()

	handler.mutex.RLock()
	clusterID := handler.clusters[0].ID
	title := handler.clusters[0].Title
	handler.mutex.RUnlock()

	code, answer := ask(t, router, `{"question":"Why is this a priority?","cluster_id":"`+clusterID+`"}`)
	if code != http.StatusOK {
		t.Fatalf("assistant returned %d, want %d", code, http.StatusOK)
	}
	if answer.Source != "offline" {
		t.Errorf("source = %q, want %q when no API key is configured", answer.Source, "offline")
	}
	if answer.Answer == "" {
		t.Fatal("assistant returned an empty answer")
	}
	if !strings.Contains(answer.Answer, title) {
		t.Errorf("answer does not reference the selected cluster %q: %s", title, answer.Answer)
	}
	if len(answer.GroundedOn) == 0 {
		t.Error("answer reports no grounding")
	}
}

// The assistant explains; it must never tell an official that something is
// approved, sanctioned or closed.
func TestAssistantNeverClaimsAuthority(t *testing.T) {
	handler, router := setupAssistantRouter()

	handler.mutex.RLock()
	clusterID := handler.clusters[0].ID
	handler.mutex.RUnlock()

	forbidden := []string{"approved", "sanctioned", "funds have been", "closed the"}
	for _, question := range []string{"Should I approve this?", "Can you close this ticket?", "Release the budget"} {
		_, answer := ask(t, router, `{"question":"`+question+`","cluster_id":"`+clusterID+`"}`)
		lowered := strings.ToLower(answer.Answer)
		for _, word := range forbidden {
			if strings.Contains(lowered, word) {
				t.Errorf("answer to %q claims authority (%q): %s", question, word, answer.Answer)
			}
		}
	}
}

// Evidence must reflect real state, including which wards are blind spots.
func TestAssistantEvidenceIncludesPlatformState(t *testing.T) {
	handler, _ := setupAssistantRouter()

	evidence, grounded := handler.buildEvidence("", "")
	if !strings.Contains(evidence, "PLATFORM OVERVIEW") {
		t.Error("evidence is missing the platform overview")
	}
	if !strings.Contains(evidence, "ALL ACTIVE CLUSTERS") {
		t.Error("evidence should list clusters when nothing is selected")
	}
	if len(grounded) == 0 {
		t.Error("evidence reports no grounding")
	}
}

func TestAssistantEvidenceQuotesCitizenReports(t *testing.T) {
	handler, _ := setupAssistantRouter()

	handler.mutex.RLock()
	cluster := handler.clusters[0]
	signal := models.CitizenSignal{ID: cluster.SignalIDs[0], Provider: models.ProviderWhatsApp, Language: "Hindi", RawText: "नल से गंदा पानी आ रहा है"}
	handler.mutex.RUnlock()

	handler.mutex.Lock()
	handler.signals = append([]models.CitizenSignal{signal}, handler.signals...)
	handler.mutex.Unlock()

	evidence, _ := handler.buildEvidence(cluster.ID, "")
	if !strings.Contains(evidence, signal.RawText) {
		t.Error("evidence does not quote the citizen report in its original language")
	}
}
