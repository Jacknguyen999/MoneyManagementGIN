.PHONY: help build up down restart logs clean test health shell db-shell db-backup dev fmt fmt-check lint

# Default target
help:
	@echo "Money Manager Backend - Docker Commands"
	@echo ""
	@echo "Available commands:"
	@echo "  make up       - Start all services"
	@echo "  make down     - Stop all services"
	@echo "  make restart  - Restart all services"
	@echo "  make build    - Build the backend image"
	@echo "  make logs     - View logs (all services)"
	@echo "  make logs-api - View backend API logs only"
	@echo "  make health   - Check service health"
	@echo "  make shell    - Access backend container shell"
	@echo "  make db-shell - Access database shell"
	@echo "  make db-migrate-large-amounts - Fix database to support large amounts"
	@echo "  make clean    - Remove containers and volumes"
	@echo "  make dev      - Start in development mode"
	@echo "  make test     - Run tests"
	@echo "  make fmt      - Format Go code"
	@echo "  make fmt-check- Check Go code formatting"
	@echo "  make lint     - Run Go linters"
	@echo ""

# # Check if Docker is running
# check-docker:
# 	@docker version > /dev/null 2>&1 || (echo "Error: Docker is not running. Please start Docker first." && exit 1)

# Check if .env exists
check-env:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "Please edit .env file with your configuration and run 'make up' again."; \
		exit 1; \
	fi

# Build the backend image
build: 
	docker-compose up --build -d

# Start all services
up: check-docker check-env
	@echo "Starting Money Manager Backend with Docker..."
	docker-compose up -d
	@echo ""
	@echo "Services started! Status:"
	@docker-compose ps
	@echo ""
	@echo "Backend API: http://localhost:8080"
	@echo "Health check: http://localhost:8080/health"
	@echo "Database: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo ""
	@echo "Run 'make logs' to view logs"
	@echo "Run 'make down' to stop services"

# Start in development mode (with logs)
dev: check-docker check-env
	@echo "Starting in development mode..."
	docker-compose up --build

# Stop all services
down:
	@echo "Stopping Money Manager Backend Docker services..."
	docker-compose down
	@echo "Services stopped!"

# Restart all services
restart: down up

# View logs for all services
logs:
	docker-compose logs -f

# View logs for backend API only
logs-api:
	docker-compose logs -f backend

# Check health of services
health:
	@echo "Checking service health..."
	@docker-compose ps
	@echo ""
	@echo "Testing backend health endpoint..."
	@curl -s http://localhost:8080/health || echo "Backend not responding"

# Access backend container shell
shell:
	docker-compose exec backend sh

# Access database shell
db-shell:
	docker-compose exec db psql -U admin -d money_manager

# Run migration for large amounts fix
db-migrate-large-amounts:
	@echo "Running migration to support large amounts..."
	docker-compose exec db psql -U admin -d money_manager -f /scripts/migrate_large_amounts.sql
	@echo "Migration completed! Database now supports amounts up to 999,999,999,999,999,999.99"

# Backup database
db-backup:
	@echo "Creating database backup..."
	docker-compose exec db pg_dump -U admin money_manager > backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created: backup_$(shell date +%Y%m%d_%H%M%S).sql"

# Clean up containers and volumes
clean:
	@echo "Removing containers and volumes..."
	docker-compose down -v
	docker system prune -f
	@echo "Cleanup complete!"

# Run tests (when implemented)
test:
	@echo "Running tests..."
	docker-compose exec backend go test ./...

# Format Go code
fmt:
	@echo "Formatting Go code..."
	@find . -name "*.go" -exec gofmt -w {} +
	@find . -name "*.go" -exec goimports -w {} +
	@echo "Go code formatted!"

# Check Go code formatting
fmt-check:
	@echo "Checking Go code formatting..."
	@gofmt -l . | grep -q . && echo "Code needs formatting. Run 'make fmt'" && exit 1 || echo "Code is properly formatted!"

# Run Go linters
lint:
	@echo "Running Go linters..."
	@golangci-lint run ./...
	@echo "Linting complete!" 