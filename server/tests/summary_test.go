package tests

import (
	"context"
	"strings"
	"testing"

	"walkover/server/internal/extraction"
	"walkover/server/internal/models"
)

func TestClusterSummary_EmptySignals(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	summary, err := ext.GenerateClusterSummary(ctx, "Ward 1 - Banganga", "Water Supply & Sewerage", []models.CitizenSignal{})
	if err != nil {
		t.Fatalf("GenerateClusterSummary failed on empty signals: %v", err)
	}

	if !strings.Contains(summary, "No active citizen signals currently recorded") {
		t.Errorf("Expected empty signal fallback notice, got: %s", summary)
	}
}

func TestClusterSummary_MultiChannelCorroboration(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	signals := []models.CitizenSignal{
		{
			ID:       "sig-01",
			Provider: models.ProviderWhatsApp,
			RawText:  "चंदन नगर में गंदा बदबूदार पानी आ रहा है सीवर का",
		},
		{
			ID:       "sig-02",
			Provider: models.ProviderWhatsApp,
			RawText:  "Street 4 Chandan Nagar tap water has black sewage smell, kids vomiting",
		},
		{
			ID:       "sig-03",
			Provider: models.ProviderSMS,
			RawText:  "Water supply contaminated in Chandan Nagar near clinic",
		},
		{
			ID:       "sig-04",
			Provider: models.ProviderWeb,
			RawText:  "Contaminated pipeline leak behind dispensary",
		},
	}

	summary, err := ext.GenerateClusterSummary(ctx, "Ward 14 - Chandan Nagar", "Water Supply & Sewerage", signals)
	if err != nil {
		t.Fatalf("GenerateClusterSummary failed: %v", err)
	}

	if !strings.Contains(summary, "Ward 14 - Chandan Nagar") {
		t.Errorf("Summary missing ward name: %s", summary)
	}

	if !strings.Contains(summary, "Water Supply & Sewerage") {
		t.Errorf("Summary missing department: %s", summary)
	}

	if !strings.Contains(summary, "4") {
		t.Errorf("Summary missing total signal count (4): %s", summary)
	}

	if !strings.Contains(summary, "WhatsApp") || !strings.Contains(summary, "SMS") {
		t.Errorf("Summary missing channel breakdown: %s", summary)
	}
}

func TestClusterSummary_HazardDetectionInSummary(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-3.6-flash", "gemini-embedding-001")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	testCases := []struct {
		name           string
		wardName       string
		department     string
		signalText     string
		expectedHazard string
	}{
		{
			name:           "Water Contamination Hazard",
			wardName:       "Ward 14 - Chandan Nagar",
			department:     "Water Supply & Sewerage",
			signalText:     "Drinking water contaminated with sewage smell",
			expectedHazard: "Water Contamination",
		},
		{
			name:           "Healthcare Route Access Hazard",
			wardName:       "Ward 22 - Vijay Nagar",
			department:     "Public Works / Roads",
			signalText:     "Ambulance stuck due to road cave in on hospital route",
			expectedHazard: "Emergency Healthcare Access Risk",
		},
		{
			name:           "Electrocution Live Wire Hazard",
			wardName:       "Ward 35 - Rajwada & Sarafa",
			department:     "Electricity & Power",
			signalText:     "Live wire sparking on main road",
			expectedHazard: "Electrocution Hazard",
		},
		{
			name:           "Open Manhole Hazard",
			wardName:       "Ward 60 - Khajrana",
			department:     "Water Supply & Sewerage",
			signalText:     "Dangerous open manhole on market street",
			expectedHazard: "Open Manhole Hazard",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signals := []models.CitizenSignal{
				{
					ID:       "sig-test",
					Provider: models.ProviderWhatsApp,
					RawText:  tc.signalText,
				},
			}

			summary, err := ext.GenerateClusterSummary(ctx, tc.wardName, tc.department, signals)
			if err != nil {
				t.Fatalf("GenerateClusterSummary failed: %v", err)
			}

			if !strings.Contains(summary, tc.expectedHazard) {
				t.Errorf("Summary missing expected hazard keyword %q. Got:\n%s", tc.expectedHazard, summary)
			}
		})
	}
}
