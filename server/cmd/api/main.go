package main

import (
	"log"

	"walkover/server/config"
	"walkover/server/internal/api"
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
