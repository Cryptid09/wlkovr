package tests

import (
	"context"
	"testing"

	"walkover/server/internal/extraction"
	"walkover/server/internal/models"
)

func BenchmarkExtractSignal_Offline(b *testing.B) {
	ctx := context.Background()
	ext, _ := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	defer ext.Close()

	signal := models.CitizenSignal{
		ID:           "bench-01",
		Provider:     models.ProviderWhatsApp,
		RawText:      "चंदन नगर गली 4 में गंदा पानी आ रहा है सीवर का",
		LocationHint: "Chandan Nagar",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ext.ExtractSignal(ctx, signal)
	}
}

func BenchmarkGenerateEmbedding_Offline(b *testing.B) {
	ctx := context.Background()
	ext, _ := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	defer ext.Close()

	text := "Vijay Nagar hospital route blocked due to road cave in"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ext.GenerateEmbedding(ctx, text)
	}
}

func BenchmarkBuildPrompt(b *testing.B) {
	ctx := context.Background()
	ext, _ := extraction.NewExtractor(ctx, "", "gemini-2.5-flash", "text-embedding-004")
	defer ext.Close()

	text := "Open manhole near Khajrana square"
	hint := "Khajrana"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ext.BuildPrompt(text, hint)
	}
}
