.PHONY: help up down restart build logs ps dev-up dev-down dev-build dev-logs dev-restart staging-up staging-down staging-build staging-logs prod-up prod-down prod-build prod-logs server-sh nginx-sh nginx-reload psql-cli db-cli clean

# Default environment: Dev / Local (Môi trường dev với local là một)
COMPOSE_DEV = docker compose -f docker-compose.yml
COMPOSE_STAGING = docker compose -f docker-compose.staging.yml
COMPOSE_PROD = docker compose -f docker-compose.product.yml

help: ## Hiển thị danh sách các lệnh
	@echo "Terra Clone - Docker & Nginx Management Commands"
	@echo "================================================="
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ----------------------------------------------------------------------
# DEV / LOCAL ENVIRONMENT (Môi trường dev với local là một)
# ----------------------------------------------------------------------
up: dev-up ## Khởi chạy môi trường Dev / Local (mặc định)
down: dev-down ## Dừng môi trường Dev / Local
build: dev-build ## Build lại image Dev / Local
logs: dev-logs ## Xem log Dev / Local
ps: dev-ps ## Kiểm tra trạng thái container Dev / Local

dev-up: ## Chạy docker compose cho Dev / Local
	$(COMPOSE_DEV) up -d

dev-down: ## Dừng docker compose cho Dev / Local
	$(COMPOSE_DEV) down

dev-build: ## Build lại container Dev / Local
	$(COMPOSE_DEV) build

dev-logs: ## Theo dõi log của tất cả service trong Dev / Local
	$(COMPOSE_DEV) logs -f

dev-ps: ## Xem danh sách container Dev / Local
	$(COMPOSE_DEV) ps

dev-restart: ## Khởi động lại các container Dev / Local
	$(COMPOSE_DEV) restart

# ----------------------------------------------------------------------
# STAGING ENVIRONMENT
# ----------------------------------------------------------------------
staging-up: ## Khởi chạy môi trường Staging
	$(COMPOSE_STAGING) up -d

staging-down: ## Dừng môi trường Staging
	$(COMPOSE_STAGING) down

staging-build: ## Build image môi trường Staging
	$(COMPOSE_STAGING) build

staging-logs: ## Theo dõi log môi trường Staging
	$(COMPOSE_STAGING) logs -f

# ----------------------------------------------------------------------
# PRODUCTION ENVIRONMENT
# ----------------------------------------------------------------------
prod-up: ## Khởi chạy môi trường Production
	$(COMPOSE_PROD) up -d

prod-down: ## Dừng môi trường Production
	$(COMPOSE_PROD) down

prod-build: ## Build image môi trường Production
	$(COMPOSE_PROD) build

prod-logs: ## Theo dõi log môi trường Production
	$(COMPOSE_PROD) logs -f

# ----------------------------------------------------------------------
# UTILITIES / DEBUGGING
# ----------------------------------------------------------------------
server-sh: ## Mở terminal sh trong container Go Server
	$(COMPOSE_DEV) exec server sh

nginx-sh: ## Mở terminal sh trong container Nginx
	$(COMPOSE_DEV) exec nginx sh

nginx-reload: ## Reload Nginx config không cần restart container
	$(COMPOSE_DEV) exec nginx nginx -s reload

psql-cli: ## Mở PostgreSQL CLI (psql) trong container postgres
	$(COMPOSE_DEV) exec postgres psql -U terra_user -d terra_db

db-cli: psql-cli ## Mở Database CLI (alias cho psql-cli)

clean: ## Dừng và xóa toàn bộ container, network, volumes của dự án
	$(COMPOSE_DEV) down -v --remove-orphans

