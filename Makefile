# ============================================================
#  RideFlow — Makefile
#  All commands for building, running, testing, and deploying
# ============================================================

# ── Variables ────────────────────────────────────────────────
MODULE       := github.com/bytepharoh/rideflow
BINARY_DIR   := bin
CMD_DIR      := cmd

SERVICES     := gateway trip driver payment

GO           := go
GOFLAGS      := -v

# ── Default target ───────────────────────────────────────────
.DEFAULT_GOAL := help

# ── Help ─────────────────────────────────────────────────────
.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Code quality ─────────────────────────────────────────────
.PHONY: fmt
fmt: ## Format all Go source files
	$(GO) fmt ./...

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet on all packages
	$(GO) vet ./...

# ── Testing ──────────────────────────────────────────────────
.PHONY: test
test: ## Run all unit tests
	$(GO) test ./... -race -count=1

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ── Build ────────────────────────────────────────────────────
.PHONY: build
build: fmt vet ## Build all service binaries into bin/
	@mkdir -p $(BINARY_DIR)
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/$$service ./$(CMD_DIR)/$$service; \
	done

.PHONY: build-service
build-service: ## Build a single service. Usage: make build-service SERVICE=trip
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/$(SERVICE) ./$(CMD_DIR)/$(SERVICE)

# ── Run (local, without Docker) ──────────────────────────────
.PHONY: run-gateway
run-gateway: ## Run the gateway service locally
	$(GO) run ./$(CMD_DIR)/gateway

.PHONY: check-gateway
check-gateway: ## Run gateway smoke checks against localhost:8080
	bash scripts/check-gateway.sh

.PHONY: run-trip
run-trip: ## Run the trip service locally
	-$(GO) run ./$(CMD_DIR)/trip

.PHONY: run-driver
run-driver: ## Run the driver service locally
	$(GO) run ./$(CMD_DIR)/driver

.PHONY: run-payment
run-payment: ## Run the payment service locally
	$(GO) run ./$(CMD_DIR)/payment

# ── Proto ────────────────────────────────────────────────────
.PHONY: proto
proto: ## Generate Go code from .proto files (requires protoc)
	@echo "Generating proto files..."
	@bash scripts/proto-gen.sh

# ── Docker ───────────────────────────────────────────────────
.PHONY: docker-build
docker-build: ## Build Docker images for all services
	@for service in $(SERVICES); do \
		echo "Building Docker image for $$service..."; \
		docker build -f deployments/docker/$$service.Dockerfile -t rideflow/$$service:latest .; \
	done

.PHONY: docker-up
docker-up: ## Start local infrastructure (MongoDB, RabbitMQ, Jaeger)
	docker compose up -d

.PHONY: docker-down
docker-down: ## Stop local infrastructure
	docker compose down

.PHONY: docker-logs
docker-logs: ## Tail logs from all running containers
	docker compose logs -f

# ── Dev (Tilt + Kubernetes) ───────────────────────────────────
.PHONY: dev
dev: ## Start local Kubernetes dev loop with Tilt
	tilt up

.PHONY: dev-down
dev-down: ## Stop Tilt dev environment
	tilt down

# ── Clean ────────────────────────────────────────────────────
.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned."	
.PHONY: push
push: ## push to github
	git push -u origin master   
.PHONY: add
add:
	git add .