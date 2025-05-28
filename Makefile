# Makefile for Database Migrations and Scripts
# Social Media Platform - MongoDB Operations

# Configuration
MONGO_URI ?= mongodb://localhost:27017
DATABASE ?= social_media_db
TIMEOUT ?= 60s

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)Database Migration and Script Commands$(NC)"
	@echo "======================================"
	@echo ""
	@echo "$(YELLOW)Configuration:$(NC)"
	@echo "  MONGO_URI: $(MONGO_URI)"
	@echo "  DATABASE:  $(DATABASE)"
	@echo "  TIMEOUT:   $(TIMEOUT)"
	@echo ""
	@echo "$(YELLOW)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make migrate-up                    # Run all pending migrations"
	@echo "  make seed                          # Seed database with test data"
	@echo "  make cleanup-dry                   # Show what cleanup would do"
	@echo "  make db-reset                      # Reset database (dangerous!)"
	@echo "  make MONGO_URI=mongodb://prod migrate-status  # Check production status"

# Migration Commands
.PHONY: migrate-up
migrate-up: ## Run all pending migrations
	@echo "$(GREEN)Running database migrations...$(NC)"
	@go run scripts/migrate.go -command=up -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -timeout="$(TIMEOUT)"

.PHONY: migrate-up-dry
migrate-up-dry: ## Show what migrations would run (dry-run)
	@echo "$(YELLOW)Dry run: Showing pending migrations...$(NC)"
	@go run scripts/migrate.go -command=up -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -dry-run

.PHONY: migrate-up-verbose
migrate-up-verbose: ## Run migrations with verbose output
	@echo "$(GREEN)Running migrations with verbose output...$(NC)"
	@go run scripts/migrate.go -command=up -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -verbose

.PHONY: migrate-status
migrate-status: ## Show migration status
	@echo "$(GREEN)Checking migration status...$(NC)"
	@go run scripts/migrate.go -command=status -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: migrate-validate
migrate-validate: ## Validate migrations
	@echo "$(GREEN)Validating migrations...$(NC)"
	@go run scripts/migrate.go -command=validate -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: migrate-down
migrate-down: ## Rollback a specific migration (requires MIGRATION variable)
	@if [ -z "$(MIGRATION)" ]; then \
		echo "$(RED)Error: MIGRATION variable is required$(NC)"; \
		echo "Usage: make migrate-down MIGRATION=001_initial_indexes"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Rolling back migration: $(MIGRATION)$(NC)"
	@go run scripts/migrate.go -command=down -migration="$(MIGRATION)" -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: migrate-create
migrate-create: ## Create a new migration (requires NAME variable)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME variable is required$(NC)"; \
		echo "Usage: make migrate-create NAME=add_user_preferences"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating new migration: $(NAME)$(NC)"
	@go run scripts/migrate.go -command=create -name="$(NAME)"

# Seeding Commands
.PHONY: seed
seed: ## Seed database with test data
	@echo "$(GREEN)Seeding database with test data...$(NC)"
	@go run scripts/seed_data.go

.PHONY: seed-small
seed-small: ## Seed database with minimal test data
	@echo "$(GREEN)Seeding database with minimal test data...$(NC)"
	@go run scripts/seed_data.go -users=10 -posts-per-user=5

.PHONY: seed-large
seed-large: ## Seed database with extensive test data
	@echo "$(GREEN)Seeding database with extensive test data...$(NC)"
	@go run scripts/seed_data.go -users=100 -posts-per-user=20

# Cleanup Commands
.PHONY: cleanup
cleanup: ## Run full database cleanup
	@echo "$(GREEN)Running database cleanup...$(NC)"
	@go run scripts/cleanup.go -operation=all -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: cleanup-dry
cleanup-dry: ## Show what cleanup would do (dry-run)
	@echo "$(YELLOW)Dry run: Showing what would be cleaned...$(NC)"
	@go run scripts/cleanup.go -operation=all -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -dry-run

.PHONY: cleanup-force
cleanup-force: ## Run cleanup without confirmation prompts
	@echo "$(GREEN)Running cleanup (forced)...$(NC)"
	@go run scripts/cleanup.go -operation=all -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -force

.PHONY: cleanup-stories
cleanup-stories: ## Clean up expired stories only
	@echo "$(GREEN)Cleaning up expired stories...$(NC)"
	@go run scripts/cleanup.go -operation=stories -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: cleanup-sessions
cleanup-sessions: ## Clean up expired sessions only
	@echo "$(GREEN)Cleaning up expired sessions...$(NC)"
	@go run scripts/cleanup.go -operation=sessions -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: cleanup-notifications
cleanup-notifications: ## Clean up expired notifications only
	@echo "$(GREEN)Cleaning up expired notifications...$(NC)"
	@go run scripts/cleanup.go -operation=notifications -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: cleanup-optimize
cleanup-optimize: ## Optimize database collections
	@echo "$(GREEN)Optimizing database collections...$(NC)"
	@go run scripts/cleanup.go -operation=optimize -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

# Database Commands
.PHONY: db-stats
db-stats: ## Show database statistics
	@echo "$(GREEN)Gathering database statistics...$(NC)"
	@go run scripts/cleanup.go -operation=stats -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)"

.PHONY: db-reset
db-reset: ## Reset database (WARNING: Destructive!)
	@echo "$(RED)WARNING: This will completely reset the database!$(NC)"
	@echo "$(RED)All data will be permanently lost!$(NC)"
	@read -p "Are you absolutely sure? [y/N]: " confirm && [ "$$confirm" = "y" ] || exit 1
	@go run scripts/migrate.go -command=reset -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -force

.PHONY: db-reset-force
db-reset-force: ## Reset database without confirmation (DANGEROUS!)
	@echo "$(RED)Force resetting database...$(NC)"
	@go run scripts/migrate.go -command=reset -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -force

# Development Commands
.PHONY: dev-setup
dev-setup: ## Set up development database (migrate + seed)
	@echo "$(GREEN)Setting up development database...$(NC)"
	@$(MAKE) migrate-up
	@$(MAKE) seed
	@echo "$(GREEN)Development database ready!$(NC)"

.PHONY: dev-reset
dev-reset: ## Reset and set up development database
	@echo "$(YELLOW)Resetting development database...$(NC)"
	@$(MAKE) db-reset-force
	@$(MAKE) dev-setup

.PHONY: dev-clean
dev-clean: ## Clean development database
	@echo "$(GREEN)Cleaning development database...$(NC)"
	@$(MAKE) cleanup-force

# Testing Commands
.PHONY: test-migrations
test-migrations: ## Test migrations (up and down)
	@echo "$(GREEN)Testing migrations...$(NC)"
	@$(MAKE) migrate-validate
	@$(MAKE) migrate-up-dry
	@echo "$(GREEN)Migration tests passed!$(NC)"

.PHONY: test-cleanup
test-cleanup: ## Test cleanup operations
	@echo "$(GREEN)Testing cleanup operations...$(NC)"
	@$(MAKE) cleanup-dry
	@echo "$(GREEN)Cleanup tests passed!$(NC)"

# Production Commands
.PHONY: prod-migrate
prod-migrate: ## Run production migrations (with confirmation)
	@echo "$(RED)WARNING: Running production migrations!$(NC)"
	@echo "Database: $(DATABASE)"
	@echo "URI: $(MONGO_URI)"
	@read -p "Continue with production migration? [y/N]: " confirm && [ "$$confirm" = "y" ] || exit 1
	@$(MAKE) migrate-validate
	@$(MAKE) migrate-up-verbose

.PHONY: prod-status
prod-status: ## Check production migration status
	@echo "$(GREEN)Checking production migration status...$(NC)"
	@$(MAKE) migrate-status

.PHONY: prod-cleanup
prod-cleanup: ## Run production cleanup (with confirmation)
	@echo "$(RED)WARNING: Running production cleanup!$(NC)"
	@echo "Database: $(DATABASE)"
	@echo "URI: $(MONGO_URI)"
	@read -p "Continue with production cleanup? [y/N]: " confirm && [ "$$confirm" = "y" ] || exit 1
	@$(MAKE) cleanup

# Maintenance Commands
.PHONY: maintenance-weekly
maintenance-weekly: ## Weekly maintenance routine
	@echo "$(GREEN)Running weekly maintenance...$(NC)"
	@$(MAKE) cleanup-stories
	@$(MAKE) cleanup-sessions
	@$(MAKE) cleanup-notifications
	@$(MAKE) db-stats

.PHONY: maintenance-monthly
maintenance-monthly: ## Monthly maintenance routine
	@echo "$(GREEN)Running monthly maintenance...$(NC)"
	@$(MAKE) cleanup-force
	@$(MAKE) cleanup-optimize
	@$(MAKE) db-stats

# Utility Commands
.PHONY: check-connection
check-connection: ## Test MongoDB connection
	@echo "$(GREEN)Testing MongoDB connection...$(NC)"
	@go run scripts/migrate.go -command=status -mongo-uri="$(MONGO_URI)" -database="$(DATABASE)" -timeout=5s

.PHONY: build-scripts
build-scripts: ## Build all scripts as binaries
	@echo "$(GREEN)Building migration scripts...$(NC)"
	@mkdir -p bin
	@go build -o bin/migrate scripts/migrate.go
	@go build -o bin/seed scripts/seed_data.go
	@go build -o bin/cleanup scripts/cleanup.go
	@echo "$(GREEN)Scripts built in bin/ directory$(NC)"

.PHONY: clean-bins
clean-bins: ## Clean built binaries
	@echo "$(GREEN)Cleaning built binaries...$(NC)"
	@rm -rf bin/

# Docker Commands (if using Docker)
.PHONY: docker-mongo
docker-mongo: ## Start MongoDB in Docker for development
	@echo "$(GREEN)Starting MongoDB in Docker...$(NC)"
	@docker run --name social-media-mongo -d -p 27017:27017 mongo:latest

.PHONY: docker-mongo-stop
docker-mongo-stop: ## Stop MongoDB Docker container
	@echo "$(GREEN)Stopping MongoDB Docker container...$(NC)"
	@docker stop social-media-mongo || true
	@docker rm social-media-mongo || true

# Backup Commands
.PHONY: backup
backup: ## Create database backup
	@echo "$(GREEN)Creating database backup...$(NC)"
	@mkdir -p backups
	@mongodump --uri="$(MONGO_URI)" --db="$(DATABASE)" --out="backups/$(DATABASE)-$(shell date +%Y%m%d-%H%M%S)"

.PHONY: restore
restore: ## Restore database from backup (requires BACKUP_DIR variable)
	@if [ -z "$(BACKUP_DIR)" ]; then \
		echo "$(RED)Error: BACKUP_DIR variable is required$(NC)"; \
		echo "Usage: make restore BACKUP_DIR=backups/social_media_db-20240101-120000"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Restoring database from: $(BACKUP_DIR)$(NC)"
	@mongorestore --uri="$(MONGO_URI)" --db="$(DATABASE)" --drop "$(BACKUP_DIR)/$(DATABASE)"

# Environment-specific targets
.PHONY: dev
dev: ## Target development environment
	$(eval MONGO_URI := mongodb://localhost:27017)
	$(eval DATABASE := social_media_dev)

.PHONY: staging
staging: ## Target staging environment
	$(eval MONGO_URI := mongodb://staging-server:27017)
	$(eval DATABASE := social_media_staging)

.PHONY: prod
prod: ## Target production environment
	$(eval MONGO_URI := mongodb://prod-cluster:27017)
	$(eval DATABASE := social_media_prod)

# Complex workflow targets
.PHONY: deploy-staging
deploy-staging: staging ## Deploy to staging environment
	@$(MAKE) prod-migrate
	@$(MAKE) cleanup-dry

.PHONY: deploy-production
deploy-production: prod ## Deploy to production environment
	@$(MAKE) backup
	@$(MAKE) prod-migrate
	@$(MAKE) prod-status

# Monitoring targets
.PHONY: monitor-migrations
monitor-migrations: ## Monitor migration progress (run in background)
	@while true; do \
		$(MAKE) migrate-status; \
		sleep 30; \
	done

.PHONY: watch-stats
watch-stats: ## Watch database statistics (run in background)
	@while true; do \
		clear; \
		$(MAKE) db-stats; \
		sleep 60; \
	done

# Validation targets
.PHONY: validate-all
validate-all: ## Run all validation checks
	@echo "$(GREEN)Running comprehensive validation...$(NC)"
	@$(MAKE) check-connection
	@$(MAKE) migrate-validate
	@$(MAKE) test-migrations
	@$(MAKE) test-cleanup
	@echo "$(GREEN)All validations passed!$(NC)"

# Emergency targets
.PHONY: emergency-rollback
emergency-rollback: ## Emergency rollback (requires MIGRATION variable)
	@if [ -z "$(MIGRATION)" ]; then \
		echo "$(RED)Error: MIGRATION variable is required$(NC)"; \
		echo "Usage: make emergency-rollback MIGRATION=002_add_social_features"; \
		exit 1; \
	fi
	@echo "$(RED)EMERGENCY ROLLBACK: $(MIGRATION)$(NC)"
	@$(MAKE) backup
	@$(MAKE) migrate-down MIGRATION="$(MIGRATION)"

.PHONY: emergency-cleanup
emergency-cleanup: ## Emergency cleanup (forced)
	@echo "$(RED)EMERGENCY CLEANUP$(NC)"
	@$(MAKE) cleanup-force

# Info targets
.PHONY: info
info: ## Show environment information
	@echo "$(GREEN)Environment Information$(NC)"
	@echo "======================="
	@echo "MongoDB URI: $(MONGO_URI)"
	@echo "Database:    $(DATABASE)"
	@echo "Timeout:     $(TIMEOUT)"
	@echo "Go Version:  $(shell go version)"
	@echo "Date:        $(shell date)"
	@echo ""
	@$(MAKE) check-connection

.PHONY: version
version: ## Show script versions and info
	@echo "$(GREEN)Database Migration System$(NC)"
	@echo "=========================="
	@echo "Version: 1.0.0"
	@echo "Platform: Social Media"
	@echo "Database: MongoDB"
	@echo "Language: Go"
	@echo ""
	@echo "Available Scripts:"
	@echo "  - migrate.go:    Migration management"
	@echo "  - seed_data.go:  Database seeding"
	@echo "  - cleanup.go:    Database cleanup"



# 	# Set up development environment
# make dev-setup              # Migrate + seed database

# # Run migrations
# make migrate-up             # Apply all pending migrations
# make migrate-status         # Check migration status

# # Database seeding
# make seed                   # Add realistic test data
# make seed-small            # Minimal test data

# # Maintenance
# make cleanup               # Clean expired data
# make cleanup-dry          # See what would be cleaned
# make db-stats             # Show database statistics

# # Production deployment
# make prod-migrate         # Production migrations (with safety)
# make prod-cleanup         # Production cleanup