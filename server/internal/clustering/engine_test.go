package clustering

import (
	"testing"

	"walkover/server/internal/models"
	"walkover/server/internal/urgency"
)

func TestCosineSimilarity(t *testing.T) {
	vecA := []float32{1.0, 0.0, 0.5}
	vecB := []float32{1.0, 0.0, 0.5}
	sim := CosineSimilarity(vecA, vecB)
	if sim < 0.999 {
		t.Errorf("Expected identical vectors to have similarity ~1.0, got %v", sim)
	}

	vecC := []float32{0.0, 1.0, 0.0}
	simOrthogonal := CosineSimilarity(vecA, vecC)
	if simOrthogonal > 0.001 {
		t.Errorf("Expected orthogonal vectors to have similarity ~0.0, got %v", simOrthogonal)
	}
}

func TestFourDimensionalScores(t *testing.T) {
	urgencyEngine := urgency.NewEngine()
	engine := NewEngine(urgencyEngine)

	scores := engine.ComputeFourDimensionalScores(
		8,    // signal count
		85.0, // urgency score
		2,    // channel count (WhatsApp + SMS)
		0.32, // poor ward infra index (Chandan Nagar)
		true, // location resolved
		true, // department mapped
	)

	if scores.Need < 80.0 {
		t.Errorf("Expected high Need score, got %v", scores.Need)
	}
	if scores.Equity < 60.0 {
		t.Errorf("Expected high Equity score for low-infra ward (0.32), got %v", scores.Equity)
	}
	if scores.Actionability != 100.0 {
		t.Errorf("Expected 100.0 Actionability for resolved ward & department, got %v", scores.Actionability)
	}
}

func TestBlindSpotDetection(t *testing.T) {
	urgencyEngine := urgency.NewEngine()
	engine := NewEngine(urgencyEngine)

	wards := []models.Ward{
		{
			ID:         "ward-poor",
			Name:       "Poor Ward",
			InfraIndex: 0.30,
		},
		{
			ID:         "ward-rich",
			Name:       "Wealthy Ward",
			InfraIndex: 0.85,
		},
	}

	clusters := []models.Cluster{
		{
			ID:     "cl-1",
			WardID: "ward-rich",
		},
	}

	evaluated := engine.DetectBlindSpots(wards, clusters)
	if !evaluated[0].IsBlindSpot {
		t.Errorf("Expected poor ward with 0 clusters to be flagged as Blind Spot")
	}
	if evaluated[1].IsBlindSpot {
		t.Errorf("Expected wealthy ward not to be flagged as Blind Spot")
	}
}

// An underserved ward that is actively reporting is visible to us, so it is a
// demand hotspot — not a blind spot. Flagging it as both contradicts the
// dashboard and the pitch.
func TestUnderservedWardWithReportsIsNotBlindSpot(t *testing.T) {
	engine := NewEngine(urgency.NewEngine())

	wards := []models.Ward{
		{ID: "ward-poor-loud", Name: "Underserved but reporting", InfraIndex: 0.32},
		{ID: "ward-poor-silent", Name: "Underserved and silent", InfraIndex: 0.35},
	}
	clusters := []models.Cluster{
		{ID: "cl-1", WardID: "ward-poor-loud"},
	}

	evaluated := engine.DetectBlindSpots(wards, clusters)
	if evaluated[0].IsBlindSpot {
		t.Errorf("Underserved ward with %d active clusters flagged as Blind Spot; want hotspot only", evaluated[0].ActiveClusterCount)
	}
	if !evaluated[1].IsBlindSpot {
		t.Errorf("Underserved ward with no reports not flagged as Blind Spot")
	}
}
