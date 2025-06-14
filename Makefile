.PHONY: help test test-backend test-frontend build start stop restart logs clean dev setup import-fixtures lint

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Testing
test: test-backend ## Run all tests
	@echo "All tests completed"

test-backend: ## Run Go backend tests
	@echo "Running backend tests..."
	cd backend && go test ./...

test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	cd frontend && npm test

# Development
dev: ## Start development environment
	docker-compose up -d

setup: ## Initial project setup (build and start)
	@echo "Setting up project..."
	docker-compose build
	docker-compose up -d
	@echo "Waiting for services to start..."
	sleep 5
	@echo "Project setup complete!"

# Docker operations
build: ## Build all Docker images
	docker-compose build

build-backend: ## Build only backend Docker image
	docker-compose build backend

build-frontend: ## Build only frontend Docker image
	docker-compose build frontend

start: ## Start all services
	docker-compose up -d

stop: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

restart-backend: ## Restart only backend service
	docker-compose restart backend

restart-frontend: ## Restart only frontend service
	docker-compose restart frontend

logs: ## Show logs for all services
	docker-compose logs -f

logs-backend: ## Show backend logs
	docker-compose logs -f backend

logs-frontend: ## Show frontend logs
	docker-compose logs -f frontend

logs-db: ## Show database logs
	docker-compose logs -f postgres

# Database operations
db-shell: ## Connect to PostgreSQL database
	docker-compose exec postgres psql -U postgres -d footy_tipping

db-reset: ## Reset database (WARNING: destroys all data)
	@echo "WARNING: This will destroy all data. Press Ctrl+C to cancel, or Enter to continue..."
	@read
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Database reset complete"

# Application operations
import-fixtures: ## Import sample match fixtures
	cd backend && ./bin/import-fixtures -db="postgres://postgres:postgres@localhost:5432/footy_tipping?sslmode=disable"

build-tools: ## Build backend tools (import-fixtures)
	cd backend && go build -o bin/import-fixtures ./cmd/import-fixtures

# Cleanup
clean: ## Clean up Docker resources
	docker-compose down -v --remove-orphans
	docker system prune -f

clean-all: ## Clean up everything including images
	docker-compose down -v --remove-orphans
	docker system prune -af

# Development helpers
fmt: ## Format Go code
	cd backend && go fmt ./...

lint: ## Run Go linter
	@echo "Running linters..."
	@cd backend && golangci-lint run

mod-tidy: ## Tidy Go modules
	cd backend && go mod tidy

# Quick development workflow
rebuild: stop build start ## Stop, rebuild, and start services

rebuild-backend: ## Rebuild and restart only backend
	docker-compose build backend
	docker-compose up -d backend

# Status checks
status: ## Show status of all services
	docker-compose ps

health: ## Check health of services
	@echo "Checking service health..."
	@echo "Backend API:"
	@curl -s http://localhost:8080/api/matches | jq 'length' || echo "Backend not responding"
	@echo "Frontend:"
	@curl -s http://localhost:3000 > /dev/null && echo "Frontend responding" || echo "Frontend not responding"
	@echo "Database:"
	@docker-compose exec postgres pg_isready -U postgres && echo "Database ready" || echo "Database not ready"

# API testing
api-test: ## Test API endpoints
	@echo "Testing API endpoints..."
	@echo "Matches endpoint:"
	curl -s http://localhost:8080/api/matches | jq 'length'
	@echo "Sample match:"
	curl -s http://localhost:8080/api/matches | jq '.[0]'
