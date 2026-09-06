package tests

import (
	"context"
	"math"
	"testing"

	"walkover/server/internal/extraction"
)

func calcCosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func TestEmbedding_Dimension768(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	samples := []string{
		"Contaminated tap water smelling like sewage in Chandan Nagar",
		"गंदा बदबूदार सीवर का पानी आ रहा है",
		"Ambulance hospital route blocked by sinkhole road cave-in",
		"Open manhole on dark road high accident risk",
		"Streetlights not working for 3 days",
	}

	for _, text := range samples {
		emb, err := ext.GenerateEmbedding(ctx, text)
		if err != nil {
			t.Fatalf("GenerateEmbedding failed for %q: %v", text, err)
		}
		if len(emb) != 768 {
			t.Errorf("Expected exactly 768 dimensions for %q, got %d", text, len(emb))
		}
	}
}

func TestEmbedding_L2Normalization(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	samples := []string{
		"Snapping high voltage live wire sparking near market",
		"Drinking water pipeline burst in Sukhliya MR-10 metro area",
		"अन्नपूर्णा मंदिर रोड पर सीवर का ढक्कन टूटा हुआ है",
	}

	for _, s := range samples {
		emb, err := ext.GenerateEmbedding(ctx, s)
		if err != nil {
			t.Fatalf("GenerateEmbedding failed: %v", err)
		}

		var sumSq float64
		for _, val := range emb {
			sumSq += float64(val * val)
		}
		magnitude := math.Sqrt(sumSq)

		if math.Abs(magnitude-1.0) > 0.05 {
			t.Errorf("Vector not L2 normalized for %q. Magnitude = %f (expected ~1.0)", s, magnitude)
		}
	}
}

func TestEmbedding_DeterministicReproducibility(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	text := "Vijay Nagar hospital road cave-in blocking ambulance route"

	emb1, err := ext.GenerateEmbedding(ctx, text)
	if err != nil {
		t.Fatalf("GenerateEmbedding 1 failed: %v", err)
	}

	emb2, err := ext.GenerateEmbedding(ctx, text)
	if err != nil {
		t.Fatalf("GenerateEmbedding 2 failed: %v", err)
	}

	for i := range emb1 {
		if emb1[i] != emb2[i] {
			t.Fatalf("Embeddings not deterministic at index %d: %f != %f", i, emb1[i], emb2[i])
		}
	}
}

func TestEmbedding_SemanticCosineClustering(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	// Two related water contamination complaints
	waterReport1 := "Contaminated brownish foul-smelling tap water in Chandan Nagar street 4"
	waterReport2 := "Drinking water smells like sewer and sewage contamination in Chandan Nagar"

	// An unrelated streetlight complaint
	streetlightReport := "Streetlight pole broken and light not functioning in colony park"

	embWater1, _ := ext.GenerateEmbedding(ctx, waterReport1)
	embWater2, _ := ext.GenerateEmbedding(ctx, waterReport2)
	embLight, _ := ext.GenerateEmbedding(ctx, streetlightReport)

	simRelated := calcCosineSimilarity(embWater1, embWater2)
	simUnrelated := calcCosineSimilarity(embWater1, embLight)

	if simRelated <= simUnrelated {
		t.Errorf("Expected related water reports to have higher similarity than unrelated reports. simRelated=%f, simUnrelated=%f",
			simRelated, simUnrelated)
	}
}

func TestEmbedding_EmptyAndWhitespaceInput(t *testing.T) {
	ctx := context.Background()
	ext, err := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	if err != nil {
		t.Fatalf("Failed to initialize extractor: %v", err)
	}
	defer ext.Close()

	testInputs := []string{"", "   ", "\t\n\r"}

	for _, input := range testInputs {
		emb, err := ext.GenerateEmbedding(ctx, input)
		if err != nil {
			t.Fatalf("GenerateEmbedding failed on empty/whitespace input %q: %v", input, err)
		}
		if len(emb) != 768 {
			t.Errorf("Expected 768 dimensions for input %q, got %d", input, len(emb))
		}
	}
}
