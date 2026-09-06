.PHONY: all dev dev-server dev-web test simulate help

help:
	@echo "Public Development Intelligence Platform (GDG Indore)"
	@echo "Available commands:"
	@echo "  make dev-server  - Run Go backend engine (port 8080)"
	@echo "  make dev-web     - Run Next.js frontend (port 3000)"
	@echo "  make test        - Run backend Go unit tests"
	@echo "  make simulate    - Trigger simulated live WhatsApp emergency message"

dev-server:
	@echo "⚡ Starting Golang Backend Engine on http://localhost:8080..."
	cd server && go run ./cmd/api/main.go

dev-web:
	@echo "🖥️ Starting Next.js Frontend Dashboard on http://localhost:3000..."
	cd web && npm run dev

test:
	@echo "🧪 Running backend unit tests..."
	cd server && go test -v ./...

simulate:
	@echo "📱 Triggering simulated live citizen message..."
	curl -X POST http://localhost:8080/api/v1/demo/simulate -H "Content-Type: application/json"
