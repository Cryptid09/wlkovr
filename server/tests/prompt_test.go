package tests

import (
	"context"
	"strings"
	"testing"

	"walkover/server/internal/extraction"
)

func TestBuildPrompt_IndoreWardCatalog(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	prompt := ext.BuildPrompt("Ganda pani aa raha hai", "Chandan Nagar")

	expectedWards := []struct {
		id   string
		name string
	}{
		{"indore-ward-01", "Ward 1 - Banganga"},
		{"indore-ward-02", "Ward 14 - Chandan Nagar"},
		{"indore-ward-03", "Ward 22 - Vijay Nagar"},
		{"indore-ward-04", "Ward 28 - Old Palasia"},
		{"indore-ward-05", "Ward 35 - Rajwada & Sarafa"},
		{"indore-ward-06", "Ward 42 - Bhawarkua & Vishnupuri"},
		{"indore-ward-07", "Ward 49 - Annapurna"},
		{"indore-ward-08", "Ward 55 - Sudama Nagar"},
		{"indore-ward-09", "Ward 60 - Khajrana"},
		{"indore-ward-10", "Ward 64 - Sukhliya"},
		{"indore-ward-11", "Ward 71 - Malharganj"},
		{"indore-ward-12", "Ward 78 - Rau & Bypass Corridor"},
	}

	for _, w := range expectedWards {
		if !strings.Contains(prompt, w.id) {
			t.Errorf("Prompt missing ward id %s", w.id)
		}
		if !strings.Contains(prompt, w.name) {
			t.Errorf("Prompt missing ward name %s", w.name)
		}
	}
}

func TestBuildPrompt_FewShotMultilingualIntegrity(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	prompt := ext.BuildPrompt("Water leak on road", "Vijay Nagar")

	// Hindi Few-Shot Verification
	if !strings.Contains(prompt, "Example 1 (Hindi)") {
		t.Error("Missing Example 1 (Hindi) in prompt")
	}
	if !strings.Contains(prompt, "चंदन नगर गली 4 में पीने के पानी में सीवेज") {
		t.Error("Missing Hindi complaint text in few-shot example")
	}

	// Hinglish Few-Shot Verification
	if !strings.Contains(prompt, "Example 2 (Hinglish)") {
		t.Error("Missing Example 2 (Hinglish) in prompt")
	}
	if !strings.Contains(prompt, "Vijay Nagar square se Mother & Child hospital") {
		t.Error("Missing Hinglish hospital complaint in few-shot example")
	}

	// English Few-Shot Verification
	if !strings.Contains(prompt, "Example 3 (English)") {
		t.Error("Missing Example 3 (English) in prompt")
	}
	if !strings.Contains(prompt, "Dangerous open manhole near Khajrana temple") {
		t.Error("Missing English manhole complaint in few-shot example")
	}

	// Hinglish Live Wire Verification
	if !strings.Contains(prompt, "Example 4 (Hinglish)") {
		t.Error("Missing Example 4 (Hinglish) in prompt")
	}
	if !strings.Contains(prompt, "LIVE_WIRE") {
		t.Error("Missing LIVE_WIRE hazard tag in few-shot example")
	}
}

func TestBuildPrompt_HazardTagsSpec(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	prompt := ext.BuildPrompt("test", "test")

	requiredHazardTags := []string{
		"CONTAMINATED_WATER",
		"LIVE_WIRE",
		"OPEN_MANHOLE",
		"HOSPITAL_ROUTE_BLOCKED",
		"ROAD_CAVE_IN",
		"SEWAGE_MIXING",
		"DRAINAGE_OVERFLOW",
	}

	for _, tag := range requiredHazardTags {
		if !strings.Contains(prompt, tag) {
			t.Errorf("Prompt missing required hazard tag specification: %s", tag)
		}
	}
}

func TestBuildPrompt_UrgencyScaleSpec(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	prompt := ext.BuildPrompt("test", "test")

	urgencyKeywords := []string{
		"5 - Immediate Critical Threat to Human Life",
		"4 - Severe Hazard / Spreading Problem",
		"3 - Moderate Civic Disruption",
		"2 - Routine Maintenance Deficit",
		"1 - Minor Cosmetic or Inconvenience",
	}

	for _, kw := range urgencyKeywords {
		if !strings.Contains(prompt, kw) {
			t.Errorf("Prompt missing urgency rating definition: %s", kw)
		}
	}
}

func TestBuildPrompt_LocationHintFormatting(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	report := "Sewage backflow in houses"
	hint := "Near Chandan Nagar Police Station Gate 2"

	prompt := ext.BuildPrompt(report, hint)

	if !strings.Contains(prompt, `Citizen Report: "Sewage backflow in houses"`) {
		t.Errorf("Citizen report not formatted properly in prompt: %s", prompt)
	}
	if !strings.Contains(prompt, `Location Hint: "Near Chandan Nagar Police Station Gate 2"`) {
		t.Errorf("Location hint not formatted properly in prompt: %s", prompt)
	}
}

func TestCleanJSONMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Pure JSON",
			input:    `{"issue": "Pothole"}`,
			expected: `{"issue": "Pothole"}`,
		},
		{
			name:     "Markdown JSON block",
			input:    "```json\n{\"issue\": \"Pothole\"}\n```",
			expected: `{"issue": "Pothole"}`,
		},
		{
			name:     "Generic Markdown block",
			input:    "```\n{\"issue\": \"Pothole\"}\n```",
			expected: `{"issue": "Pothole"}`,
		},
		{
			name:     "Whitespace padded",
			input:    "   \n```json\n{\"issue\": \"Pothole\"}\n```   \n",
			expected: `{"issue": "Pothole"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extraction.CleanJSONMarkdown(tc.input)
			if got != tc.expected {
				t.Errorf("CleanJSONMarkdown(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}
