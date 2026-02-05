.PHONY: all build run clean test frontend-dev frontend-build backend-dev

# Variablen
BACKEND_BIN := backend/bin/server
GO := go
NPM := npm

all: build

# Backend
build:
	cd backend && $(GO) build -o bin/server ./cmd/server

run: build
	./$(BACKEND_BIN)

backend-dev:
	cd backend && $(GO) run ./cmd/server

backend-test:
	cd backend && $(GO) test ./...

backend-deps:
	cd backend && $(GO) mod download && $(GO) mod tidy

# Frontend
frontend-dev:
	cd frontend && $(NPM) run dev

frontend-build:
	cd frontend && $(NPM) run build

frontend-deps:
	cd frontend && $(NPM) install

# Development
dev:
	@echo "Starte Backend und Frontend parallel..."
	@make -j2 backend-dev frontend-dev

# Datenbank
db-reset:
	rm -f data/sc2analytics.db
	mkdir -p data

# Clean
clean:
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf data/uploads/temp/*

# Hilfe
help:
	@echo "SC2 Analytics - Verfügbare Befehle:"
	@echo ""
	@echo "  make build        - Backend kompilieren"
	@echo "  make run          - Backend starten (kompiliert erst)"
	@echo "  make backend-dev  - Backend im Dev-Modus starten"
	@echo "  make backend-deps - Go Dependencies installieren"
	@echo ""
	@echo "  make frontend-dev   - Frontend Dev-Server starten"
	@echo "  make frontend-build - Frontend für Production bauen"
	@echo "  make frontend-deps  - npm Dependencies installieren"
	@echo ""
	@echo "  make dev          - Backend + Frontend parallel starten"
	@echo "  make clean        - Build-Artefakte löschen"
	@echo "  make db-reset     - Datenbank zurücksetzen"
