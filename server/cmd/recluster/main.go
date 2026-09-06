package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"walkover/server/config"
	"walkover/server/internal/api"
	"walkover/server/internal/db"
	"walkover/server/internal/extraction"
)

func main() {
	apply := flag.Bool("apply", false, "apply the live-only assignment repair")
	flag.Parse()
	ctx := context.Background()
	cfg := config.LoadConfig()
	repo, err := db.NewRepository(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()
	hub := api.NewHub()
	go hub.Run()
	handler := api.NewHandler(cfg, hub)
	if err := handler.WithPersistence(ctx, repo); err != nil {
		log.Fatal(err)
	}
	extractor, err := extraction.NewExtractor(ctx, "", cfg.GeminiModel, cfg.EmbeddingModel)
	if err != nil {
		log.Fatal(err)
	}
	defer extractor.Close()
	handler.WithExtractor(extractor)

	preview := handler.PreviewLiveAssignmentRepair()
	fmt.Printf("live signals: %d, incompatible/missing assignments: %d\n", preview.LiveSignals, preview.IncorrectMatches)
	if !*apply {
		fmt.Println("dry run only; rerun with --apply after reviewing the result")
		return
	}
	if repo.Backend() == db.BackendLocalJSON {
		backup, err := backupLocalStore()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("backup:", backup)
	}
	report, err := handler.RepairLiveAssignments(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("repair complete: %+v\n", report)
}

func backupLocalStore() (string, error) {
	var source string
	for _, candidate := range []string{"data/local_store", "../data/local_store", "server/data/local_store"} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			source = candidate
			break
		}
	}
	if source == "" {
		return "", fmt.Errorf("local store directory not found")
	}
	destination := source + ".backup-" + time.Now().Format("20060102-150405")
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, contents, 0o644)
	})
	return destination, err
}
