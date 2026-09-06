package extraction

import (
	"context"
	"testing"

	"walkover/server/internal/models"
)

// These tests guard the invariant that let clustering fail silently in
// production: the offline fallback emitted 768-dimensional vectors while the
// live model emitted 3072, and CosineSimilarity scores mismatched lengths as
// 0. Nothing errored — signals simply stopped matching any cluster.

func TestEmbeddingDimensionsPerModel(t *testing.T) {
	cases := map[string]int{
		"gemini-embedding-001": 3072,
		"text-embedding-004":   768,
		"embedding-001":        768,
		"some-future-model":    3072,
	}

	for model, want := range cases {
		if got := embeddingDimensions(model); got != want {
			t.Errorf("embeddingDimensions(%q) = %d, want %d", model, got, want)
		}
	}
}

// Every vector an extractor produces — from either code path — must have the
// width its configured model produces, or vectors become incomparable.
func TestFallbackWidthMatchesConfiguredModel(t *testing.T) {
	ctx := context.Background()

	for _, model := range []string{"gemini-embedding-001", "text-embedding-004"} {
		want := embeddingDimensions(model)

		// An empty API key forces the offline fallback path.
		extractor, err := NewExtractor(ctx, "", "gemini-3.6-flash", model)
		if err != nil {
			t.Fatalf("NewExtractor(%q): %v", model, err)
		}
		if !extractor.IsOffline() {
			t.Fatalf("extractor for %q should be offline without an API key", model)
		}

		vector, err := extractor.GenerateEmbedding(ctx, "खजराना में मैनहोल खुला पड़ा है")
		if err != nil {
			t.Fatalf("GenerateEmbedding(%q): %v", model, err)
		}
		if len(vector) != want {
			t.Errorf("GenerateEmbedding with %q returned %d dims, want %d", model, len(vector), want)
		}

		extraction, err := extractor.ExtractSignal(ctx, models.CitizenSignal{
			ID:      "sig-test-0001",
			RawText: "Khajrana me manhole ka dhakkan khula hai",
		})
		if err != nil {
			t.Fatalf("ExtractSignal(%q): %v", model, err)
		}
		if len(extraction.Embedding) != want {
			t.Errorf("ExtractSignal with %q returned %d dims, want %d", model, len(extraction.Embedding), want)
		}

		// The two paths must agree, or a signal embedded during extraction
		// could never be compared with one embedded directly.
		if len(vector) != len(extraction.Embedding) {
			t.Errorf("model %q: GenerateEmbedding gives %d dims but ExtractSignal gives %d",
				model, len(vector), len(extraction.Embedding))
		}
	}
}

// An extractor built without the dimension field set must still report a
// sensible width rather than producing zero-length vectors.
func TestDimensionsToleratesUnsetField(t *testing.T) {
	extractor := &Extractor{embeddingModel: "gemini-embedding-001"}
	if got := extractor.dimensions(); got != 3072 {
		t.Errorf("dimensions() with unset embeddingDim = %d, want 3072", got)
	}
}
