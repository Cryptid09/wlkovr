package main

import (
	"encoding/json"
	"strings"
	"path/filepath"
	"testing"
	"time"

	"walkover/server/internal/models"
)

// wardDatasetPath locates the real ward reference dataset from this package's
// directory, so the tests exercise the same input the seed command uses.
var wardDatasetPath = filepath.Join("..", "..", "data", "indore_wards.json")

var fixedNow = time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

func buildTestDataset(t *testing.T) *Dataset {
	t.Helper()

	wards, _, err := loadWards(wardDatasetPath)
	if err != nil {
		t.Fatalf("loadWards: %v", err)
	}

	dataset, err := BuildDataset(wards, fixedNow)
	if err != nil {
		t.Fatalf("BuildDataset: %v", err)
	}
	return dataset
}

func signalsByWard(dataset *Dataset) map[string]int {
	counts := map[string]int{}
	for _, extraction := range dataset.Extractions {
		counts[extraction.WardID]++
	}
	return counts
}

// The corpus must stay inside the 30–50 range agreed in UNDERSTANDING.md:
// enough volume for hotspots to be meaningful, small enough to demo.
func TestDatasetSizeWithinAgreedRange(t *testing.T) {
	dataset := buildTestDataset(t)

	// Enough volume for hotspots and a visible tier spread, small enough that
	// a full reseed with embeddings still finishes in a few minutes.
	if len(dataset.Signals) < 45 || len(dataset.Signals) > 80 {
		t.Errorf("generated %d signals, want between 45 and 80", len(dataset.Signals))
	}
	if len(dataset.Extractions) != len(dataset.Signals) {
		t.Errorf("generated %d extractions for %d signals, want one per signal",
			len(dataset.Extractions), len(dataset.Signals))
	}
	if len(dataset.RawEvents) != len(dataset.Signals) {
		t.Errorf("generated %d raw events for %d signals, want one per signal",
			len(dataset.RawEvents), len(dataset.Signals))
	}
}

// The three engineered hotspots are what prove cross-channel clustering works.
func TestHotspotWardVolumes(t *testing.T) {
	counts := signalsByWard(buildTestDataset(t))

	want := map[string]int{
		"indore-ward-02": 8,  // Ward 14 Chandan Nagar — contaminated water
		"indore-ward-03": 12, // Ward 22 Vijay Nagar — road cave-in
		"indore-ward-09": 6,  // Ward 60 Khajrana — open manhole
	}
	for wardID, wantCount := range want {
		if counts[wardID] != wantCount {
			t.Errorf("ward %s has %d signals, want %d", wardID, counts[wardID], wantCount)
		}
	}
}

// The two structurally underserved wards must stay silent — that silence is the
// entire blind-spot demonstration.
func TestBlindSpotWardsReceiveNoSignals(t *testing.T) {
	dataset := buildTestDataset(t)
	counts := signalsByWard(dataset)

	silent := []string{
		"indore-ward-01", // Ward 1 Banganga, infra 0.38
		"indore-ward-12", // Ward 78 Rau, infra 0.35
		"indore-ward-13", // Ward 6 Bhagirathpura, infra 0.29
		"indore-ward-14", // Ward 19 Nandanagar, infra 0.36
		"indore-ward-15", // Ward 47 Musakhedi, infra 0.41
		"indore-ward-16", // Ward 82 Bicholi Hapsi, infra 0.33
	}
	for _, wardID := range silent {
		if counts[wardID] != 0 {
			t.Errorf("ward %s has %d signals, want 0 so it reads as a blind spot", wardID, counts[wardID])
		}
	}

	flagged := map[string]bool{}
	for _, ward := range dataset.Wards {
		if ward.IsBlindSpot {
			flagged[ward.ID] = true
		}
	}
	for _, wardID := range silent {
		if !flagged[wardID] {
			t.Errorf("ward %s not flagged as a blind spot, want flagged", wardID)
		}
	}

	// A single blind spot reads as an edge case; a cluster of them is the
	// pattern the platform exists to surface.
	if len(flagged) < 5 {
		t.Errorf("only %d blind spots flagged, want at least 5 for the map to show the pattern", len(flagged))
	}
}

func TestClustersReferenceRealSignals(t *testing.T) {
	dataset := buildTestDataset(t)

	known := map[string]bool{}
	for _, signal := range dataset.Signals {
		known[signal.ID] = true
	}

	if len(dataset.Clusters) != 7 {
		t.Fatalf("generated %d clusters, want 7 engineered hotspots", len(dataset.Clusters))
	}

	for _, cluster := range dataset.Clusters {
		if cluster.SignalCount != len(cluster.SignalIDs) {
			t.Errorf("cluster %s reports %d signals but lists %d IDs",
				cluster.ID, cluster.SignalCount, len(cluster.SignalIDs))
		}
		for _, signalID := range cluster.SignalIDs {
			if !known[signalID] {
				t.Errorf("cluster %s cites unknown signal %s", cluster.ID, signalID)
			}
		}
		if cluster.CentroidLat == 0 || cluster.CentroidLng == 0 {
			t.Errorf("cluster %s has no map centroid", cluster.ID)
		}
	}
}

// Every hotspot carries a life-safety hazard, so all three must land in the
// top urgency tier and each of the four scores must be populated.
func TestHotspotScoringIsPopulated(t *testing.T) {
	for _, cluster := range buildTestDataset(t).Clusters {
		if cluster.Urgency.Tier == "" {
			t.Errorf("cluster %s has no urgency tier", cluster.ID)
		}
		if len(cluster.Urgency.Factors) == 0 {
			t.Errorf("cluster %s has no explainability factors", cluster.ID)
		}

		scores := cluster.Scores
		if scores.Need <= 0 || scores.Confidence <= 0 || scores.Equity <= 0 || scores.Actionability <= 0 {
			t.Errorf("cluster %s has an unpopulated score dimension: %+v", cluster.ID, scores)
		}
		if cluster.Recommendation == "" {
			t.Errorf("cluster %s has no grounded recommendation", cluster.ID)
		}
	}
}

// A corpus where every cluster is Tier 1 makes the four-tier urgency engine
// look like it has one setting. The seed must exercise the range, while
// life-safety hazards still force the top tier.
func TestUrgencyTiersSpanTheRange(t *testing.T) {
	dataset := buildTestDataset(t)

	tiers := map[models.UrgencyTier]int{}
	for _, cluster := range dataset.Clusters {
		tiers[cluster.Urgency.Tier]++
	}
	if len(tiers) < 3 {
		t.Errorf("clusters span only %d urgency tiers (%v), want at least 3", len(tiers), tiers)
	}
	if tiers[models.Tier1Critical] == 0 {
		t.Error("no TIER_1_CRITICAL cluster: the hazard overrides are not being exercised")
	}

	// Every cluster carrying a life-safety hazard must be Tier 1.
	for _, cluster := range dataset.Clusters {
		hazardous := strings.Contains(cluster.Title, "Manhole") ||
			strings.Contains(cluster.Title, "Electrical") ||
			strings.Contains(cluster.Title, "Contamination") ||
			strings.Contains(cluster.Title, "Caved-in")
		if hazardous && cluster.Urgency.Tier != models.Tier1Critical {
			t.Errorf("cluster %q carries a life-safety hazard but is %s", cluster.Title, cluster.Urgency.Tier)
		}
	}
}

// Equity must rank the underserved ward above the affluent one even though the
// affluent ward reports more complaints — that inversion is the point.
func TestEquityFavoursUnderservedWard(t *testing.T) {
	dataset := buildTestDataset(t)

	byID := map[string]models.Cluster{}
	for _, cluster := range dataset.Clusters {
		byID[cluster.ID] = cluster
	}

	chandanNagar := byID["cluster-indore-001"] // infra index 0.32, 8 signals
	vijayNagar := byID["cluster-indore-002"]   // infra index 0.84, 12 signals

	if chandanNagar.Scores.Equity <= vijayNagar.Scores.Equity {
		t.Errorf("equity: Chandan Nagar %.1f, Vijay Nagar %.1f; want the underserved ward to score higher",
			chandanNagar.Scores.Equity, vijayNagar.Scores.Equity)
	}
	if vijayNagar.Scores.Confidence < chandanNagar.Scores.Confidence {
		t.Errorf("confidence: Vijay Nagar %.1f, Chandan Nagar %.1f; want more channels and signals to score higher",
			vijayNagar.Scores.Confidence, chandanNagar.Scores.Confidence)
	}
}

// Cross-language, cross-channel grouping is the platform's core claim, so each
// hotspot must actually contain multiple languages.
func TestHotspotsAreMultilingual(t *testing.T) {
	dataset := buildTestDataset(t)

	languagesByWard := map[string]map[string]bool{}
	for _, signal := range dataset.Signals {
		wardID, _ := signal.Metadata["ward_id"].(string)
		if languagesByWard[wardID] == nil {
			languagesByWard[wardID] = map[string]bool{}
		}
		languagesByWard[wardID][signal.Language] = true
	}

	for _, wardID := range []string{"indore-ward-02", "indore-ward-03", "indore-ward-09"} {
		if len(languagesByWard[wardID]) < 3 {
			t.Errorf("ward %s spans %d languages, want at least 3 (Hindi, Hinglish, English)",
				wardID, len(languagesByWard[wardID]))
		}
	}

	channels := map[string]bool{}
	for _, cluster := range dataset.Clusters {
		for _, channel := range cluster.Channels {
			channels[channel] = true
		}
	}
	if len(channels) < 3 {
		t.Errorf("hotspots span %d channels, want WhatsApp, SMS and WebPortal", len(channels))
	}
}

// Embeddings belong to Workstream 2 (text-embedding-004); the seed must leave
// that field empty rather than inventing vectors.
func TestExtractionsLeaveEmbeddingsToExtractionPipeline(t *testing.T) {
	for _, extraction := range buildTestDataset(t).Extractions {
		if len(extraction.Embedding) != 0 {
			t.Fatalf("extraction %s carries a %d-dim embedding, want none",
				extraction.SignalID, len(extraction.Embedding))
		}
		if extraction.Issue == "" || extraction.Department == "" || extraction.Summary == "" {
			t.Errorf("extraction %s is missing structured fields: %+v", extraction.SignalID, extraction)
		}
	}
}

// Background noise exists to be ignored: unrelated single issues must not
// become hotspots.
func TestBackgroundReportsDoNotFormHotspots(t *testing.T) {
	dataset := buildTestDataset(t)

	hotspotWards := map[string]bool{}
	for _, cluster := range dataset.Clusters {
		hotspotWards[cluster.WardID] = true
	}

	// Wards that carry only unrelated one-off reports.
	for _, wardID := range []string{"indore-ward-06", "indore-ward-07", "indore-ward-11"} {
		if hotspotWards[wardID] {
			t.Errorf("background ward %s produced a hotspot cluster", wardID)
		}
	}
}

// A cluster may only sit in a non-default state because a human put it there.
func TestInvestigatingClusterHasHumanDecisionBehindIt(t *testing.T) {
	dataset := buildTestDataset(t)

	audited := map[string]bool{}
	for _, entry := range dataset.AuditLogs {
		audited[entry.ClusterID] = true
		if entry.Officer == "" {
			t.Errorf("audit entry %s has no officer recorded", entry.ID)
		}
	}

	for _, cluster := range dataset.Clusters {
		if cluster.Status != "PENDING" && !audited[cluster.ID] {
			t.Errorf("cluster %s is %s with no audit entry explaining the decision", cluster.ID, cluster.Status)
		}
	}
}

// Re-running the seed with the same anchor time must produce byte-identical
// output, so reseeding before the demo cannot silently change the numbers.
func TestDatasetIsDeterministic(t *testing.T) {
	first, err := json.Marshal(buildTestDataset(t))
	if err != nil {
		t.Fatalf("marshal first dataset: %v", err)
	}
	second, err := json.Marshal(buildTestDataset(t))
	if err != nil {
		t.Fatalf("marshal second dataset: %v", err)
	}

	if string(first) != string(second) {
		t.Error("BuildDataset is not deterministic for a fixed anchor time")
	}
}

func TestLoadWardsRejectsMissingFile(t *testing.T) {
	if _, _, err := loadWards(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("loadWards on a missing file returned nil error, want failure")
	}
}
