# Go Integration - Makefile
# Comandos para build, teste e deploy

# Configurações
BINARY_NAME=go-integration
API_BINARY=api
WORKER_BINARY=worker
BIN_DIR=bin

# Variáveis de build
GO_VERSION := $(shell go version | cut -d ' ' -f 3)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X 'main.Version=${GIT_COMMIT}' -X 'main.BuildTime=${BUILD_TIME}'

# Cores para output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help build build-api build-worker clean test run-api run-worker docker-up docker-down start stop restart dev

help: ## Mostrar ajuda
	@echo "$(GREEN)Go Integration - Comandos Disponíveis:$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: build-api build-worker ## Compilar todos os binários

build-api: ## Compilar API REST
	@echo "$(GREEN)🔨 Compilando API...$(NC)"
	@go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(API_BINARY) ./cmd/api
	@echo "$(GREEN)✅ API compilada: $(BIN_DIR)/$(API_BINARY)$(NC)"

build-worker: ## Compilar Worker
	@echo "$(GREEN)🔨 Compilando Worker...$(NC)"
	@go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(WORKER_BINARY) ./cmd/worker
	@echo "$(GREEN)✅ Worker compilado: $(BIN_DIR)/$(WORKER_BINARY)$(NC)"

clean: ## Limpar binários e cache
	@echo "$(YELLOW)🧹 Limpando binários...$(NC)"
	@rm -rf $(BIN_DIR)/*
	@go clean -cache
	@echo "$(GREEN)✅ Limpeza concluída$(NC)"

test: ## Executar testes
	@echo "$(GREEN)🧪 Executando testes...$(NC)"
	@go test -v ./...

run-api: build-api ## Compilar e executar API
	@echo "$(GREEN)🚀 Executando API...$(NC)"
	@./$(BIN_DIR)/$(API_BINARY)

run-worker: build-worker ## Compilar e executar Worker
	@echo "$(GREEN)🚀 Executando Worker...$(NC)"
	@./$(BIN_DIR)/$(WORKER_BINARY)

dev-api: ## Executar API em modo desenvolvimento (sem compilar)
	@echo "$(GREEN)🔧 Executando API em modo dev...$(NC)"
	@go run ./cmd/api

dev-worker: ## Executar Worker em modo desenvolvimento (sem compilar)
	@echo "$(GREEN)🔧 Executando Worker em modo dev...$(NC)"
	@go run ./cmd/worker

docker-up: ## Iniciar infraestrutura Docker
	@echo "$(GREEN)🐳 Iniciando infraestrutura...$(NC)"
	@docker-compose -f .infra/docker-compose.dev.yml up -d
	@echo "$(GREEN)✅ Infraestrutura iniciada$(NC)"

docker-down: ## Parar infraestrutura Docker
	@echo "$(YELLOW)🐳 Parando infraestrutura...$(NC)"
	@docker-compose -f .infra/docker-compose.dev.yml down
	@echo "$(GREEN)✅ Infraestrutura parada$(NC)"

install: ## Instalar dependências
	@echo "$(GREEN)📦 Instalando dependências...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependências instaladas$(NC)"

setup: install docker-up build ## Setup completo do projeto
	@echo ""
	@echo "$(GREEN)🎉 Setup completo!$(NC)"
	@echo ""
	@echo "$(YELLOW)Próximos passos:$(NC)"
	@echo "  1. Configure RESEND_API_KEY no arquivo .env"
	@echo "  2. Execute: make run-worker"
	@echo "  3. Execute: make run-api (terminal separado)"
	@echo ""

info: ## Mostrar informações do projeto
	@echo "$(GREEN)📋 Informações do Projeto:$(NC)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Binários: $(BIN_DIR)/"
	@ls -la $(BIN_DIR)/ 2>/dev/null || echo "  (nenhum binário compilado)"

# Comandos de desenvolvimento
watch-api: ## Observar mudanças na API (requer 'air')
	@which air > /dev/null || (echo "$(RED)❌ 'air' não instalado. Execute: go install github.com/cosmtrek/air@latest$(NC)" && exit 1)
	@air -c .air-api.toml

watch-worker: ## Observar mudanças no Worker (requer 'air')
	@which air > /dev/null || (echo "$(RED)❌ 'air' não instalado. Execute: go install github.com/cosmtrek/air@latest$(NC)" && exit 1)
	@air -c .air-worker.toml
