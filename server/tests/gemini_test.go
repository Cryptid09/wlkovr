package tests

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"walkover/server/internal/extraction"
	"walkover/server/internal/models"
)

type MultilingualFixture struct {
	ID                 string   `json:"id"`
	Language           string   `json:"language"`
	RawText            string   `json:"raw_text"`
	ExpectedWardID     string   `json:"expected_ward_id"`
	ExpectedWardName   string   `json:"expected_ward_name"`
	ExpectedDepartment string   `json:"expected_department"`
	ExpectedHazardTags []string `json:"expected_hazard_tags"`
	MinUrgency         int      `json:"min_urgency"`
}

func loadMultilingualFixtures(t *testing.T) []MultilingualFixture {
	data, err := os.ReadFile("fixtures/raw_complaints_multilingual.json")
	if err != nil {
		data, err = os.ReadFile("../internal/extraction/testdata/raw_complaints_multilingual.json")
	}
	if err != nil {
		t.Fatalf("Failed to read test fixtures: %v", err)
	}

	var fixtures []MultilingualFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("Failed to parse test fixtures: %v", err)
	}
	return fixtures
}

func TestExtractSignal_AllMultilingualFixtures(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	fixtures := loadMultilingualFixtures(t)
	if len(fixtures) < 5 {
		t.Fatalf("Expected at least 5 fixtures, got %d", len(fixtures))
	}

	for _, f := range fixtures {
		t.Run(f.ID+"_"+f.Language, func(t *testing.T) {
			signal := models.CitizenSignal{
				ID:       f.ID,
				Provider: models.ProviderWhatsApp,
				RawText:  f.RawText,
			}

			extractionResult, err := ext.ExtractSignal(ctx, signal)
			if err != nil {
				t.Fatalf("ExtractSignal failed for %s: %v", f.ID, err)
			}

			if extractionResult.Issue == "" {
				t.Errorf("Extraction issue is empty for %s", f.ID)
			}

			if extractionResult.WardID != f.ExpectedWardID {
				t.Errorf("WardID mismatch for %s: expected %s, got %s", f.ID, f.ExpectedWardID, extractionResult.WardID)
			}

			if extractionResult.Department != f.ExpectedDepartment {
				t.Errorf("Department mismatch for %s: expected %s, got %s", f.ID, f.ExpectedDepartment, extractionResult.Department)
			}

			if extractionResult.BaseUrgency < f.MinUrgency {
				t.Errorf("BaseUrgency too low for %s: expected >= %d, got %d", f.ID, f.MinUrgency, extractionResult.BaseUrgency)
			}

			for _, expectedTag := range f.ExpectedHazardTags {
				found := false
				for _, tag := range extractionResult.HazardTags {
					if tag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Missing expected hazard tag %s for %s. Got: %v", expectedTag, f.ID, extractionResult.HazardTags)
				}
			}

			if len(extractionResult.Embedding) != 3072 {
				t.Errorf("Expected 3072-dim embedding, got %d", len(extractionResult.Embedding))
			}

			if extractionResult.ConfidenceScore <= 0 || extractionResult.ConfidenceScore > 1.0 {
				t.Errorf("Invalid confidence score %f", extractionResult.ConfidenceScore)
			}
		})
	}
}
