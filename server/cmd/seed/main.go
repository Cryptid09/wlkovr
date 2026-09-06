// Command seed populates the platform's collections with the engineered demo
// corpus: three multilingual demand hotspots, two silent blind-spot wards, and
// unrelated background reports.
//
// It writes to Firestore when a Firestore emulator host or a Google service
// account credential is configured, and to local JSON files otherwise, so the
// corpus is reproducible with or without Google Cloud access.
//
//	go run ./cmd/seed              # seed using whichever backend is configured
//	go run ./cmd/seed --reset      # clear all collections first
//	go run ./cmd/seed --dry-run    # build and summarise without writing
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"walkover/server/config"
	"walkover/server/internal/db"
	"walkover/server/internal/extraction"
	"walkover/server/internal/models"
)

func main() {
	var (
		wardsPath = flag.String("wards", "", "path to the ward reference dataset (default: search data/indore_wards.json)")
		localDir  = flag.String("local-dir", "", "force the local JSON backend and write collections into this directory")
		reset     = flag.Bool("reset", false, "delete every document in every collection before seeding")
		dryRun    = flag.Bool("dry-run", false, "build and summarise the dataset without writing anything")
		embed     = flag.Bool("embed", false, "after seeding, generate Gemini embeddings for every extraction that lacks one")
	)
	flag.Parse()

	if err := run(*wardsPath, *localDir, *reset, *dryRun, *embed); err != nil {
		log.Fatalf("[SEED] %v", err)
	}
}

func run(wardsPath, localDir string, reset, dryRun, embed bool) error {
	wards, resolvedPath, err := loadWards(wardsPath)
	if err != nil {
		return err
	}
	log.Printf("[SEED] Loaded %d wards from %s", len(wards), resolvedPath)

	dataset, err := BuildDataset(wards, time.Now())
	if err != nil {
		return err
	}
	summarise(dataset)

	if dryRun {
		log.Printf("[SEED] Dry run — nothing written")
		return nil
	}

	ctx := context.Background()
	repo, err := openRepository(ctx, localDir)
	if err != nil {
		return err
	}
	defer repo.Close()

	log.Printf("[SEED] Backend: %s", repo.Backend())

	if reset {
		if err := repo.Reset(ctx); err != nil {
			return fmt.Errorf("resetting collections: %w", err)
		}
		log.Printf("[SEED] Cleared all collections")
	}

	if err := write(ctx, repo, dataset); err != nil {
		return err
	}

	if embed {
		if err := backfillEmbeddings(ctx, repo); err != nil {
			return fmt.Errorf("backfilling embeddings: %w", err)
		}
	}

	log.Printf("[SEED] Done — %d signals across %d wards, %d hotspot clusters",
		len(dataset.Signals), len(dataset.Wards), len(dataset.Clusters))
	return nil
}

// backfillEmbeddings gives seeded extractions the vectors the seed itself does
// not generate, so live signals can be matched against historical evidence by
// semantic similarity rather than by department label alone.
func backfillEmbeddings(ctx context.Context, repo db.Repository) error {
	cfg := config.LoadConfig()

	extractor, err := extraction.NewExtractor(ctx, cfg.GeminiAPIKey, cfg.GeminiModel, cfg.EmbeddingModel)
	if err != nil {
		return err
	}
	defer extractor.Close()

	if extractor.IsOffline() {
		log.Printf("[SEED] GEMINI_API_KEY not set — skipping embedding backfill")
		return nil
	}

	extractions, err := repo.ListExtractions(ctx, 0)
	if err != nil {
		return err
	}

	embedded, failed := 0, 0
	for _, item := range extractions {
		if len(item.Embedding) > 0 {
			continue
		}

		vector, err := extractor.GenerateEmbedding(ctx, db.EmbeddingText(item))
		if err != nil {
			// One bad embedding must not abort the whole backfill; the signal
			// simply falls back to department matching.
			log.Printf("[SEED] embedding %s: %v", item.SignalID, err)
			failed++
			continue
		}

		item.Embedding = vector
		if err := repo.SaveExtraction(ctx, item); err != nil {
			return err
		}
		embedded++
	}

	log.Printf("[SEED] Embedded %d extractions (%d failed) using %s", embedded, failed, cfg.EmbeddingModel)
	return nil
}

// openRepository honours an explicit local directory, otherwise lets the
// configuration decide between Firestore and the local JSON fallback.
func openRepository(ctx context.Context, localDir string) (db.Repository, error) {
	if localDir != "" {
		return db.NewLocalRepository(localDir)
	}

	cfg := config.LoadConfig()
	repo, err := db.NewRepository(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("opening repository: %w", err)
	}
	if repo.Backend() == db.BackendLocalJSON {
		log.Printf("[SEED] No Firestore emulator host or service account credential configured — using local JSON store")
	}
	return repo, nil
}

// write persists the dataset collection by collection.
func write(ctx context.Context, repo db.Repository, dataset *Dataset) error {
	if err := repo.UpsertWards(ctx, dataset.Wards); err != nil {
		return fmt.Errorf("writing wards: %w", err)
	}
	for _, event := range dataset.RawEvents {
		if err := repo.SaveRawEvent(ctx, event); err != nil {
			return fmt.Errorf("writing raw events: %w", err)
		}
	}
	for _, signal := range dataset.Signals {
		if err := repo.SaveSignal(ctx, signal); err != nil {
			return fmt.Errorf("writing signals: %w", err)
		}
	}
	for _, extraction := range dataset.Extractions {
		if err := repo.SaveExtraction(ctx, extraction); err != nil {
			return fmt.Errorf("writing extractions: %w", err)
		}
	}
	for _, cluster := range dataset.Clusters {
		if err := repo.UpsertCluster(ctx, cluster); err != nil {
			return fmt.Errorf("writing clusters: %w", err)
		}
	}
	for _, hotspot := range dataset.Hotspots {
		if err := repo.UpsertHotspot(ctx, hotspot); err != nil {
			return fmt.Errorf("writing hotspots: %w", err)
		}
	}
	for _, rec := range dataset.Recommendations {
		if err := repo.SaveRecommendation(ctx, rec); err != nil {
			return fmt.Errorf("writing recommendations: %w", err)
		}
	}
	for _, entry := range dataset.AuditLogs {
		if err := repo.AppendAuditLog(ctx, entry); err != nil {
			return fmt.Errorf("writing audit logs: %w", err)
		}
	}
	return nil
}

// summarise prints what the corpus proves, so a failed demo assumption is
// visible before the data is written rather than during the pitch.
func summarise(dataset *Dataset) {
	log.Printf("[SEED] Built %d signals, %d extractions, %d clusters, %d recommendations, %d audit entries",
		len(dataset.Signals), len(dataset.Extractions), len(dataset.Clusters),
		len(dataset.Recommendations), len(dataset.AuditLogs))

	for _, cluster := range dataset.Clusters {
		log.Printf("[SEED]   hotspot %s | %s | %d signals via %v | urgency %.1f (%s) | need %.1f confidence %.1f equity %.1f actionability %.1f",
			cluster.ID, cluster.WardName, cluster.SignalCount, cluster.Channels,
			cluster.Urgency.Score, cluster.Urgency.Tier,
			cluster.Scores.Need, cluster.Scores.Confidence, cluster.Scores.Equity, cluster.Scores.Actionability)
	}

	for _, ward := range dataset.Wards {
		if ward.IsBlindSpot {
			log.Printf("[SEED]   blind spot %s | %s | infra index %.2f | %d active clusters",
				ward.ID, ward.Name, ward.InfraIndex, ward.ActiveClusterCount)
		}
	}
}

// loadWards reads the ward reference dataset, searching the usual locations so
// the command works from the repo root or from server/.
func loadWards(path string) ([]models.Ward, string, error) {
	candidates := []string{path}
	if path == "" {
		candidates = []string{
			filepath.Join("data", "indore_wards.json"),
			filepath.Join("..", "data", "indore_wards.json"),
			filepath.Join("server", "data", "indore_wards.json"),
		}
	}

	for _, candidate := range candidates {
		raw, err := os.ReadFile(candidate)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, candidate, fmt.Errorf("reading wards from %s: %w", candidate, err)
		}

		var wards []models.Ward
		if err := json.Unmarshal(raw, &wards); err != nil {
			return nil, candidate, fmt.Errorf("parsing wards from %s: %w", candidate, err)
		}
		if len(wards) == 0 {
			return nil, candidate, fmt.Errorf("ward dataset %s is empty", candidate)
		}
		return wards, candidate, nil
	}

	return nil, "", fmt.Errorf("ward dataset not found (looked in %v)", candidates)
}
