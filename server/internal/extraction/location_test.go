package extraction

import (
	"testing"

	"walkover/server/internal/models"
)

func TestResolveWardFromHinglishVariant(t *testing.T) {
	ward := resolveWard("Khajrane ke yaha manhole khula hai", "", "", "")
	if ward.ID != "indore-ward-09" || ward.Confidence < 0.8 {
		t.Fatalf("unexpected resolution: %+v", ward)
	}
}

func TestResolveWardUsesValidatedModelCandidate(t *testing.T) {
	ward := resolveWard("road toot gayi hai", "", "indore-ward-03", "")
	if ward.ID != "indore-ward-03" || ward.Source != "model_inferred" {
		t.Fatalf("unexpected model resolution: %+v", ward)
	}
}

func TestResolveWardWithoutLocationIsLowConfidenceAndNotBanganga(t *testing.T) {
	ward := resolveWard("road toot gayi hai", "", "invalid", "invented")
	if ward.ID == "indore-ward-01" || ward.Confidence >= 0.6 {
		t.Fatalf("unsafe fallback resolution: %+v", ward)
	}
}

func TestFireCategoryIsCanonical(t *testing.T) {
	category := canonicalCategory("", "Fire & Emergency Services", "Active fire", []string{"FIRE"})
	if category != models.CategoryFire {
		t.Fatalf("got %s", category)
	}
}
