package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"walkover/server/internal/models"
)

// newTestRepo returns a local JSON repository rooted in a temp directory, so
// the whole suite runs without Google Cloud credentials or an emulator.
func newTestRepo(t *testing.T) Repository {
	t.Helper()

	repo, err := NewLocalRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalRepository: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	return repo
}

func TestLocalRepositoryBackend(t *testing.T) {
	if got := newTestRepo(t).Backend(); got != BackendLocalJSON {
		t.Errorf("Backend() = %q, want %q", got, BackendLocalJSON)
	}
}

func TestSignalRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	sent := time.Date(2026, 9, 6, 10, 30, 0, 0, time.UTC)
	signal := models.CitizenSignal{
		ID:           "seed-sig-0001",
		Provider:     models.ProviderWhatsApp,
		RawText:      "नल से गंदा पानी आ रहा है",
		Language:     "Hindi",
		LocationHint: "Chandan Nagar Street 4",
		Timestamp:    sent,
	}

	if err := repo.SaveSignal(ctx, signal); err != nil {
		t.Fatalf("SaveSignal: %v", err)
	}

	got, err := repo.ListSignals(ctx, 0)
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListSignals returned %d signals, want 1", len(got))
	}

	// Devanagari text and timestamps must survive the JSON round trip intact.
	if got[0].RawText != signal.RawText {
		t.Errorf("RawText = %q, want %q", got[0].RawText, signal.RawText)
	}
	if !got[0].Timestamp.Equal(sent) {
		t.Errorf("Timestamp = %v, want %v", got[0].Timestamp, sent)
	}
}

func TestSaveRejectsEmptyID(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	if err := repo.SaveSignal(ctx, models.CitizenSignal{RawText: "no id"}); err == nil {
		t.Fatal("SaveSignal with empty ID returned nil error, want failure")
	}
}

func TestUpsertReplacesDocument(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	cluster := models.Cluster{ID: "cluster-indore-001", Status: "PENDING", SignalCount: 8}
	if err := repo.UpsertCluster(ctx, cluster); err != nil {
		t.Fatalf("UpsertCluster: %v", err)
	}

	cluster.Status = "INVESTIGATING"
	if err := repo.UpsertCluster(ctx, cluster); err != nil {
		t.Fatalf("UpsertCluster (update): %v", err)
	}

	clusters, err := repo.ListClusters(ctx)
	if err != nil {
		t.Fatalf("ListClusters: %v", err)
	}
	if len(clusters) != 1 {
		t.Fatalf("ListClusters returned %d clusters, want 1 after upsert", len(clusters))
	}
	if clusters[0].Status != "INVESTIGATING" {
		t.Errorf("Status = %q, want %q", clusters[0].Status, "INVESTIGATING")
	}
}

func TestGetClusterNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	if _, err := repo.GetCluster(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetCluster error = %v, want ErrNotFound", err)
	}
}

func TestListOrdersByDocumentIDAndAppliesLimit(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	// Inserted out of order to prove listing sorts rather than preserving
	// insertion order — the Firestore backend orders by document ID too.
	for _, id := range []string{"seed-sig-0003", "seed-sig-0001", "seed-sig-0002"} {
		if err := repo.SaveSignal(ctx, models.CitizenSignal{ID: id}); err != nil {
			t.Fatalf("SaveSignal(%s): %v", id, err)
		}
	}

	all, err := repo.ListSignals(ctx, 0)
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	want := []string{"seed-sig-0001", "seed-sig-0002", "seed-sig-0003"}
	for i, id := range want {
		if all[i].ID != id {
			t.Errorf("signal[%d].ID = %q, want %q", i, all[i].ID, id)
		}
	}

	limited, err := repo.ListSignals(ctx, 2)
	if err != nil {
		t.Fatalf("ListSignals(limit 2): %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("ListSignals(limit 2) returned %d signals, want 2", len(limited))
	}
}

func TestExtractionKeepsEmbeddingEmpty(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	// Workstream 2 owns embeddings; the data layer must not invent them.
	if err := repo.SaveExtraction(ctx, models.AIExtraction{SignalID: "seed-sig-0001", Issue: "Contaminated drinking water supply"}); err != nil {
		t.Fatalf("SaveExtraction: %v", err)
	}

	extractions, err := repo.ListExtractions(ctx, 0)
	if err != nil {
		t.Fatalf("ListExtractions: %v", err)
	}
	if len(extractions) != 1 {
		t.Fatalf("ListExtractions returned %d extractions, want 1", len(extractions))
	}
	if extractions[0].Embedding != nil {
		t.Errorf("Embedding = %v, want nil", extractions[0].Embedding)
	}
}

func TestEveryCollectionPersists(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	if err := repo.UpsertWards(ctx, []models.Ward{{ID: "indore-ward-01", Name: "Ward 1 - Banganga"}}); err != nil {
		t.Fatalf("UpsertWards: %v", err)
	}
	if err := repo.SaveRawEvent(ctx, RawEvent{ID: "seed-evt-0001", SignalID: "seed-sig-0001"}); err != nil {
		t.Fatalf("SaveRawEvent: %v", err)
	}
	if err := repo.SaveSignal(ctx, models.CitizenSignal{ID: "seed-sig-0001"}); err != nil {
		t.Fatalf("SaveSignal: %v", err)
	}
	if err := repo.SaveExtraction(ctx, models.AIExtraction{SignalID: "seed-sig-0001"}); err != nil {
		t.Fatalf("SaveExtraction: %v", err)
	}
	if err := repo.UpsertCluster(ctx, models.Cluster{ID: "cluster-indore-001"}); err != nil {
		t.Fatalf("UpsertCluster: %v", err)
	}
	if err := repo.UpsertHotspot(ctx, models.Cluster{ID: "cluster-indore-001"}); err != nil {
		t.Fatalf("UpsertHotspot: %v", err)
	}
	if err := repo.SaveRecommendation(ctx, Recommendation{ID: "seed-rec-001", ClusterID: "cluster-indore-001"}); err != nil {
		t.Fatalf("SaveRecommendation: %v", err)
	}
	if err := repo.AppendAuditLog(ctx, models.AuditLog{ID: "seed-audit-001", Action: "INVESTIGATE"}); err != nil {
		t.Fatalf("AppendAuditLog: %v", err)
	}

	wards, err := repo.ListWards(ctx)
	if err != nil || len(wards) != 1 {
		t.Errorf("ListWards = %d docs, err %v; want 1 doc", len(wards), err)
	}
	rawEvents, err := repo.ListRawEvents(ctx, 0)
	if err != nil || len(rawEvents) != 1 {
		t.Errorf("ListRawEvents = %d docs, err %v; want 1 doc", len(rawEvents), err)
	}
	hotspots, err := repo.ListHotspots(ctx)
	if err != nil || len(hotspots) != 1 {
		t.Errorf("ListHotspots = %d docs, err %v; want 1 doc", len(hotspots), err)
	}
	recommendations, err := repo.ListRecommendations(ctx, 0)
	if err != nil || len(recommendations) != 1 {
		t.Errorf("ListRecommendations = %d docs, err %v; want 1 doc", len(recommendations), err)
	}
	auditLogs, err := repo.ListAuditLogs(ctx, 0)
	if err != nil || len(auditLogs) != 1 {
		t.Errorf("ListAuditLogs = %d docs, err %v; want 1 doc", len(auditLogs), err)
	}
}

func TestResetClearsAllCollections(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	if err := repo.SaveSignal(ctx, models.CitizenSignal{ID: "seed-sig-0001"}); err != nil {
		t.Fatalf("SaveSignal: %v", err)
	}
	if err := repo.UpsertCluster(ctx, models.Cluster{ID: "cluster-indore-001"}); err != nil {
		t.Fatalf("UpsertCluster: %v", err)
	}

	if err := repo.Reset(ctx); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	signals, err := repo.ListSignals(ctx, 0)
	if err != nil {
		t.Fatalf("ListSignals after reset: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("ListSignals after reset returned %d signals, want 0", len(signals))
	}

	clusters, err := repo.ListClusters(ctx)
	if err != nil {
		t.Fatalf("ListClusters after reset: %v", err)
	}
	if len(clusters) != 0 {
		t.Errorf("ListClusters after reset returned %d clusters, want 0", len(clusters))
	}
}

// Reset on an untouched store must be a no-op rather than an error, so
// `--reset` works on a first run.
func TestResetOnEmptyStore(t *testing.T) {
	if err := newTestRepo(t).Reset(context.Background()); err != nil {
		t.Errorf("Reset on empty store: %v", err)
	}
}
