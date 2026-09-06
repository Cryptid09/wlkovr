package tests

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"walkover/server/internal/extraction"
	"walkover/server/internal/models"
)

type FixtureItem struct {
	ID                 string   `json:"id"`
	Language           string   `json:"language"`
	RawText            string   `json:"raw_text"`
	ExpectedWardID     string   `json:"expected_ward_id"`
	ExpectedWardName   string   `json:"expected_ward_name"`
	ExpectedDepartment string   `json:"expected_department"`
	ExpectedHazardTags []string `json:"expected_hazard_tags"`
	MinUrgency         int      `json:"min_urgency"`
}

func loadIntegrationFixtures(t *testing.T) []FixtureItem {
	data, err := os.ReadFile("fixtures/raw_complaints_multilingual.json")
	if err != nil {
		data, err = os.ReadFile("../internal/extraction/testdata/raw_complaints_multilingual.json")
	}
	if err != nil {
		t.Fatalf("Failed to read test fixtures: %v", err)
	}

	var fixtures []FixtureItem
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("Failed to parse test fixtures: %v", err)
	}
	return fixtures
}

func TestExtraction_EndToEndConsumerPipeline(t *testing.T) {
	ctx := context.Background()
	extractor, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize Extractor: %v", err)
	}
	defer extractor.Close()

	fixtures := loadIntegrationFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("No fixtures loaded")
	}

	var extractedSignals []models.CitizenSignal

	for _, fix := range fixtures {
		signal := models.CitizenSignal{
			ID:          fix.ID,
			Provider:    models.ProviderWhatsApp,
			RawText:     fix.RawText,
			SenderPhone: "+91-9876543210",
			Timestamp:   time.Now(),
		}

		extractionResult, err := extractor.ExtractSignal(ctx, signal)
		if err != nil {
			t.Fatalf("ExtractSignal failed for %s: %v", fix.ID, err)
		}

		if extractionResult.SignalID != fix.ID {
			t.Errorf("SignalID mismatch: expected %s, got %s", fix.ID, extractionResult.SignalID)
		}

		if extractionResult.WardID != fix.ExpectedWardID {
			t.Errorf("WardID mismatch for %s: expected %s, got %s", fix.ID, fix.ExpectedWardID, extractionResult.WardID)
		}

		if extractionResult.Department != fix.ExpectedDepartment {
			t.Errorf("Department mismatch for %s: expected %s, got %s", fix.ID, fix.ExpectedDepartment, extractionResult.Department)
		}

		if len(extractionResult.Embedding) != 3072 {
			t.Errorf("Embedding dimension mismatch for %s: expected 3072, got %d", fix.ID, len(extractionResult.Embedding))
		}

		extractedSignals = append(extractedSignals, signal)
	}

	// Test grounded cluster summary on collected signals
	summary, err := extractor.GenerateClusterSummary(ctx, "Ward 14 - Chandan Nagar", "Water Supply & Sewerage", extractedSignals)
	if err != nil {
		t.Fatalf("GenerateClusterSummary failed: %v", err)
	}

	if !strings.Contains(summary, "Ward 14 - Chandan Nagar") {
		t.Errorf("Summary missing ward name: %s", summary)
	}
}

func TestExtraction_PromptIntegrityContract(t *testing.T) {
	ctx := context.Background()
	extractor, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize Extractor: %v", err)
	}
	defer extractor.Close()

	prompt := extractor.BuildPrompt("Ganda pani aa raha hai", "Chandan Nagar")

	// Ensure system prompt is non-empty and well-structured
	if len(prompt) < 500 {
		t.Errorf("System prompt too short (%d bytes)", len(prompt))
	}

	if !strings.Contains(prompt, "indore-ward-01") {
		t.Error("System prompt missing Ward 1 mapping")
	}
}
