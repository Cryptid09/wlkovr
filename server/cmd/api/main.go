package main

import (
	"context"
	"log"

	"walkover/server/config"
	"walkover/server/internal/api"
	"walkover/server/internal/db"
	"walkover/server/internal/extraction"
)

func main() {
	log.Println("=========================================================")
	log.Println("⚡ Starting Public Development Intelligence Engine (Go)")
	log.Println("=========================================================")

	// 1. Load Central Configuration
	cfg := config.LoadConfig()
	log.Printf("[CONFIG] Environment: %s | Port: %s", cfg.Env, cfg.Port)

	// 2. Initialize WebSocket Hub
	hub := api.NewHub()
	go hub.Run()
	log.Println("[WS] Real-time WebSocket Broadcaster Hub initialized")

	// 3. Initialize API Handlers & In-Memory Store
	handler := api.NewHandler(cfg, hub)
	log.Println("[ENGINE] Urgency Decision Engine & In-Memory Clustering initialized")

	ctx := context.Background()

	// 3a. Attach persistence. The repository picks Firestore when an emulator
	// host or service account credential is configured, and a local JSON store
	// otherwise, so the server runs either way.
	repo, err := db.NewRepository(ctx, cfg)
	if err != nil {
		log.Printf("[STORE] Persistence unavailable, serving in-memory only: %v", err)
	} else {
		defer repo.Close()
		if err := handler.WithPersistence(ctx, repo); err != nil {
			log.Printf("[STORE] Hydration failed, serving in-memory only: %v", err)
		}
	}

	// 3b. Attach the Gemini extraction pipeline. It degrades to deterministic
	// offline fallbacks when no API key is present.
	extractor, err := extraction.NewExtractor(ctx, cfg.GeminiAPIKey, cfg.GeminiModel, cfg.EmbeddingModel)
	if err != nil {
		log.Printf("[GEMINI] Extractor unavailable, signals will be ingested without understanding: %v", err)
	} else {
		defer extractor.Close()
		handler.WithExtractor(extractor)
	}

	// 4. Setup Gin Router
	router := api.SetupRouter(cfg, handler, hub)

	// 5. Start HTTP & WebSocket Server
	serverAddr := ":" + cfg.Port
	log.Printf("[HTTP] Server listening on http://localhost%s", serverAddr)
	log.Printf("[API]  Endpoints live at http://localhost%s/api/v1", serverAddr)
	log.Printf("[WS]   WebSocket live at ws://localhost%s/ws", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("[FATAL] Failed to start server: %v", err)
	}
}
