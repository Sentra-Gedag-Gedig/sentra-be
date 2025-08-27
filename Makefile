# Sentra Backend Makefile
# Supports both local development and production deployment

include .env
export

# Configuration
DOCKER_REGISTRY ?= docker.io
DOCKER_USERNAME ?= your-username
IMAGE_NAME = sentra-backend
IMAGE_TAG ?= latest
DEV_COMPOSE_FILE = docker-compose.dev.yml
PROD_COMPOSE_FILE = docker-compose.prod.yml

# Image names
DEV_IMAGE_NAME = $(DOCKER_USERNAME)/$(IMAGE_NAME):dev
PROD_IMAGE_NAME = $(DOCKER_USERNAME)/$(IMAGE_NAME):$(IMAGE_TAG)
REGISTRY_IMAGE_NAME = $(DOCKER_REGISTRY)/$(DOCKER_USERNAME)/$(IMAGE_NAME):$(IMAGE_TAG)

# Colors for output
BLUE = \033[36m
GREEN = \033[32m  
YELLOW = \033[33m
RED = \033[31m
NC = \033[0m # No Color

.PHONY: help setup clean

# Default target
help: ## Show this help message
	@echo "$(BLUE)Sentra Backend - Development & Production Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "$(BLUE)%-20s$(NC) %s\n", $$1, $$2}' | \
		sort
	@echo ""
	@echo "$(YELLOW)Environment Variables:$(NC)"
	@echo "  DOCKER_REGISTRY   = $(DOCKER_REGISTRY)"
	@echo "  DOCKER_USERNAME   = $(DOCKER_USERNAME)"
	@echo "  IMAGE_TAG         = $(IMAGE_TAG)"
	@echo "  DEV_IMAGE         = $(DEV_IMAGE_NAME)"  
	@echo "  PROD_IMAGE        = $(PROD_IMAGE_NAME)"
	@echo ""
	@echo "$(YELLOW)Usage Examples:$(NC)"
	@echo "  make dev-start              # Start development environment"
	@echo "  make prod-build-push        # Build and push production image"
	@echo "  make deploy                 # Full deployment pipeline"

setup: ## Setup development environment
	@echo "$(BLUE)🔧 Setting up development environment...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)📝 Creating .env from .env.example$(NC)"; \
		cp .env.example .env; \
		echo "$(RED)⚠️  Please edit .env with your configuration$(NC)"; \
	fi
	@mkdir -p storage/logs
	@mkdir -p nginx/logs
	@mkdir -p postgres
	@mkdir -p redis
	@echo "$(GREEN)✅ Setup completed$(NC)"

# =================
# Development Commands
# =================

dev-build: setup ## Build development image
	@echo "$(BLUE)🔨 Building development image...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) build golang-app
	@echo "$(GREEN)✅ Development image built$(NC)"

dev-start: setup ## Start development environment
	@echo "$(BLUE)🚀 Starting development environment...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) up -d
	@echo "$(GREEN)✅ Development environment started$(NC)"
	@echo "$(YELLOW)📍 Application: http://localhost:$(APP_PORT)$(NC)"
	@echo "$(YELLOW)📍 Database: localhost:5432$(NC)"
	@echo "$(YELLOW)📍 Redis: localhost:6379$(NC)"

dev-start-with-tools: setup ## Start development environment with additional tools
	@echo "$(BLUE)🚀 Starting development environment with tools...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) --profile minio --profile mailhog --profile redis-insight --profile pgadmin up -d
	@echo "$(GREEN)✅ Development environment with tools started$(NC)"
	@echo "$(YELLOW)📍 Application: http://localhost:$(APP_PORT)$(NC)"
	@echo "$(YELLOW)📍 pgAdmin: http://localhost:8080 (admin@sentra.dev/admin123)$(NC)"
	@echo "$(YELLOW)📍 Redis Insight: http://localhost:8001$(NC)"
	@echo "$(YELLOW)📍 MinIO: http://localhost:9001 (minioadmin/minioadmin123)$(NC)"
	@echo "$(YELLOW)📍 MailHog: http://localhost:8025$(NC)"

dev-logs: ## Show development logs
	@docker compose -f $(DEV_COMPOSE_FILE) logs -f golang-app

dev-shell: ## Open shell in development container
	@docker compose -f $(DEV_COMPOSE_FILE) exec golang-app sh

dev-test: dev-build ## Run tests in development environment
	@echo "$(BLUE)🧪 Running tests in development environment...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) up -d
	@sleep 10
	@docker compose -f $(DEV_COMPOSE_FILE) exec golang-app go test -v ./...
	@echo "$(GREEN)✅ Tests completed$(NC)"

dev-stop: ## Stop development environment
	@echo "$(BLUE)🛑 Stopping development environment...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) down
	@echo "$(GREEN)✅ Development environment stopped$(NC)"

dev-clean: ## Clean development environment
	@echo "$(BLUE)🧹 Cleaning development environment...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) down --volumes --remove-orphans
	@docker image prune -f
	@echo "$(GREEN)✅ Development environment cleaned$(NC)"

dev-restart: dev-stop dev-start ## Restart development environment

# =================
# Production Commands  
# =================

prod-build: setup ## Build production image
	@echo "$(BLUE)🏗️ Building production image...$(NC)"
	@docker build --target production -t $(PROD_IMAGE_NAME) .
	@if [ "$(DOCKER_REGISTRY)" != "docker.io" ]; then \
		docker tag $(PROD_IMAGE_NAME) $(REGISTRY_IMAGE_NAME); \
	fi
	@echo "$(GREEN)✅ Production image built: $(PROD_IMAGE_NAME)$(NC)"

prod-build-slim: setup ## Build slim production image (scratch-based)
	@echo "$(BLUE)🏗️ Building slim production image...$(NC)"
	@docker build --target production-slim -t $(PROD_IMAGE_NAME)-slim .
	@echo "$(GREEN)✅ Slim production image built: $(PROD_IMAGE_NAME)-slim$(NC)"

prod-test: prod-build ## Test production image locally
	@echo "$(BLUE)🧪 Testing production image locally...$(NC)"
	@docker run --rm -d --name sentra-test \
		-p 8081:8080 \
		-e DB_HOST=host.docker.internal \
		-e REDIS_ADDRESS=host.docker.internal:6379 \
		$(PROD_IMAGE_NAME)
	@sleep 10
	@if curl -f http://localhost:8081/ > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Production image test passed$(NC)"; \
	else \
		echo "$(RED)❌ Production image test failed$(NC)"; \
		docker logs sentra-test; \
	fi
	@docker stop sentra-test || true
	@docker rm sentra-test || true

prod-push: ## Push production image to registry  
	@echo "$(BLUE)📤 Pushing production image to registry...$(NC)"
	@if [ "$(DOCKER_REGISTRY)" = "docker.io" ]; then \
		docker login; \
		docker push $(PROD_IMAGE_NAME); \
	else \
		docker login $(DOCKER_REGISTRY); \
		docker push $(REGISTRY_IMAGE_NAME); \
	fi
	@echo "$(GREEN)✅ Production image pushed$(NC)"

prod-build-push: prod-build prod-push ## Build and push production image

prod-deploy: ## Deploy production environment (pull from registry)
	@echo "$(BLUE)🚀 Deploying production environment...$(NC)"
	@docker compose -f $(PROD_COMPOSE_FILE) pull
	@docker compose -f $(PROD_COMPOSE_FILE) up -d
	@echo "$(GREEN)✅ Production environment deployed$(NC)"
	@echo "$(YELLOW)📍 Application: http://localhost$(NC)"

prod-logs: ## Show production logs
	@docker compose -f $(PROD_COMPOSE_FILE) logs -f golang-app

prod-shell: ## Open shell in production container  
	@docker compose -f $(PROD_COMPOSE_FILE) exec golang-app sh

prod-stop: ## Stop production environment
	@echo "$(BLUE)🛑 Stopping production environment...$(NC)"
	@docker compose -f $(PROD_COMPOSE_FILE) down
	@echo "$(GREEN)✅ Production environment stopped$(NC)"

prod-restart: prod-stop prod-deploy ## Restart production environment

# =================
# Database Commands
# =================

db-migrate-up: ## Run database migrations (development)
	@echo "$(BLUE)📈 Running database migrations...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) exec golang-app make migrate-up
	@echo "$(GREEN)✅ Migrations completed$(NC)"

db-migrate-down: ## Rollback database migrations (development)
	@echo "$(BLUE)📉 Rolling back database migrations...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) exec golang-app make migrate-down
	@echo "$(GREEN)✅ Rollback completed$(NC)"

db-shell: ## Open PostgreSQL shell (development)
	@docker compose -f $(DEV_COMPOSE_FILE) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

# =================  
# Monitoring Commands
# =================

monitor-start: ## Start monitoring stack (production)
	@echo "$(BLUE)📊 Starting monitoring stack...$(NC)"
	@docker compose -f $(PROD_COMPOSE_FILE) --profile monitoring up -d
	@echo "$(GREEN)✅ Monitoring stack started$(NC)"
	@echo "$(YELLOW)📍 Prometheus: http://localhost:9090$(NC)"
	@echo "$(YELLOW)📍 Grafana: http://localhost:3000$(NC)"

logs-start: ## Start log aggregation (production)
	@echo "$(BLUE)📝 Starting log aggregation...$(NC)"
	@docker compose -f $(PROD_COMPOSE_FILE) --profile logging up -d
	@echo "$(GREEN)✅ Log aggregation started$(NC)"

# =================
# Utility Commands
# =================

ps: ## Show running containers
	@echo "$(BLUE)Development Containers:$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) ps 2>/dev/null || echo "No development containers running"
	@echo ""
	@echo "$(BLUE)Production Containers:$(NC)"  
	@docker compose -f $(PROD_COMPOSE_FILE) ps 2>/dev/null || echo "No production containers running"

images: ## Show built images
	@echo "$(BLUE)Local Images:$(NC)"
	@docker images | grep -E "(sentra-backend|$(DOCKER_USERNAME))" | head -10

clean: ## Clean up Docker resources
	@echo "$(BLUE)🧹 Cleaning up Docker resources...$(NC)"
	@docker compose -f $(DEV_COMPOSE_FILE) down --volumes --remove-orphans 2>/dev/null || true
	@docker compose -f $(PROD_COMPOSE_FILE) down --volumes --remove-orphans 2>/dev/null || true
	@docker image prune -f
	@docker volume prune -f
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

health: ## Check services health
	@echo "$(BLUE)🏥 Checking services health...$(NC)"
	@echo "Development:"
	@docker compose -f $(DEV_COMPOSE_FILE) ps --format "table {{.Service}}\t{{.State}}\t{{.Health}}" 2>/dev/null || echo "  No services running"
	@echo ""
	@echo "Production:"
	@docker compose -f $(PROD_COMPOSE_FILE) ps --format "table {{.Service}}\t{{.State}}\t{{.Health}}" 2>/dev/null || echo "  No services running"

# =================
# CI/CD Commands
# =================

ci-test: ## Run CI tests
	@echo "$(BLUE)🔄 Running CI tests...$(NC)"
	@docker build --target development -t sentra-test .
	@docker run --rm sentra-test go test -v ./...
	@echo "$(GREEN)✅ CI tests passed$(NC)"

deploy: setup dev-test prod-build prod-test prod-push ## Full deployment pipeline
	@echo ""
	@echo "$(GREEN)🎉 Deployment pipeline completed successfully!$(NC)"
	@echo ""
	@echo "$(YELLOW)📋 Next steps:$(NC)"
	@echo "  1. Copy these files to your server:"
	@echo "     - $(PROD_COMPOSE_FILE)" 
	@echo "     - .env (with production values)"
	@echo "     - nginx/"
	@echo "     - private.key"
	@echo ""
	@echo "  2. On your server, run:"
	@echo "     $(BLUE)make prod-deploy$(NC)"
	@echo ""
	@echo "  3. Optional: Start monitoring:"
	@echo "     $(BLUE)make monitor-start$(NC)"

# =================
# Version Commands
# =================

version: ## Show version information
	@echo "$(BLUE)Sentra Backend Version Information$(NC)"
	@echo "Docker Compose Dev: $(DEV_COMPOSE_FILE)"
	@echo "Docker Compose Prod: $(PROD_COMPOSE_FILE)"  
	@echo "Image Tag: $(IMAGE_TAG)"
	@docker --version
	@docker compose version