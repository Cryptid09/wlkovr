package urgency

import (
	"testing"
	"time"

	"walkover/server/internal/models"
)

func TestUrgencyDecisionEngine_HardHazardOverride(t *testing.T) {
	engine := NewEngine()

	// 1 citizen reports live wire with low base urgency (1/5)
	res := engine.CalculateUrgency(
		1,
		[]string{"LIVE_WIRE"},
		1,
		1,
		[]string{},
		time.Now(),
	)

	if res.Tier != models.Tier1Critical {
		t.Errorf("Expected Tier 1 Critical for LIVE_WIRE hazard, got %v", res.Tier)
	}
	if res.Score < 85.0 {
		t.Errorf("Expected score >= 85.0 for LIVE_WIRE hazard, got %v", res.Score)
	}
	if res.SLAHours != 4 {
		t.Errorf("Expected 4 hour SLA for Tier 1 Critical, got %d", res.SLAHours)
	}
}

func TestUrgencyDecisionEngine_VelocitySpike(t *testing.T) {
	engine := NewEngine()

	// 6 complaints within 2 hours
	res := engine.CalculateUrgency(
		3,
		[]string{},
		6,
		6,
		[]string{},
		time.Now(),
	)

	if res.Score < 70.0 {
		t.Errorf("Expected score >= 70.0 with velocity surge, got %v", res.Score)
	}
	if res.VelocityRate != 6 {
		t.Errorf("Expected velocity rate 6, got %v", res.VelocityRate)
	}
}
