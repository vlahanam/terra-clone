.PHONY: help up down restart build logs ps dev-up dev-down dev-build dev-logs dev-restart staging-up staging-down staging-build staging-logs prod-up prod-down prod-build prod-logs server-sh nginx-sh nginx-reload psql-cli db-cli clean migration

# [VN] Default environment: Dev / Local
COMPOSE_DEV = docker compose -f docker-compose.yml
COMPOSE_STAGING = docker compose -f docker-compose.staging.yml
COMPOSE_PROD = docker compose -f docker-compose.product.yml

help: ## [VN] Hiển thị danh sách các lệnh
	@echo "Terra Clone - Docker & Nginx Management Commands"
	@echo "================================================="
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ----------------------------------------------------------------------
# DEV / LOCAL ENVIRONMENT
# ----------------------------------------------------------------------
up: dev-up ## [VN] Khởi chạy môi trường Dev / Local (mặc định)[VN]
down: dev-down ## [VN] Dừng môi trường Dev / Local
build: dev-build ## [VN] Build lại image Dev / Local
logs: dev-logs ## [VN] Xem log Dev / Local
ps: dev-ps ## [VN] Kiểm tra trạng thái container Dev / Local

dev-up: ## [VN] Chạy docker compose cho Dev / Local
	$(COMPOSE_DEV) up -d

dev-down: ## [VN] Dừng docker compose cho Dev / Local
	$(COMPOSE_DEV) down

dev-build: ## [VN] Build lại container Dev / Local
	$(COMPOSE_DEV) build

dev-logs: ## [VN] Theo dõi log của tất cả service trong Dev / Local
	$(COMPOSE_DEV) logs -f

dev-ps: ## [VN] Xem danh sách container Dev / Local
	$(COMPOSE_DEV) ps

dev-restart: ## [VN] Khởi động lại các container Dev / Local
	$(COMPOSE_DEV) restart

# ----------------------------------------------------------------------
# STAGING ENVIRONMENT
# ----------------------------------------------------------------------
staging-up: ## [VN] Khởi chạy môi trường Staging
	$(COMPOSE_STAGING) up -d

staging-down: ## [VN] Dừng môi trường Staging
	$(COMPOSE_STAGING) down

staging-build: ## [VN] Build image môi trường Staging
	$(COMPOSE_STAGING) build

staging-logs: ## [VN] Theo dõi log môi trường Staging
	$(COMPOSE_STAGING) logs -f

# ----------------------------------------------------------------------
# PRODUCTION ENVIRONMENT
# ----------------------------------------------------------------------
prod-up: ## [VN] Khởi chạy môi trường Production
	$(COMPOSE_PROD) up -d

prod-down: ## [VN] Dừng môi trường Production
	$(COMPOSE_PROD) down

prod-build: ## [VN] Build image môi trường Production
	$(COMPOSE_PROD) build

prod-logs: ## [VN] Theo dõi log môi trường Production
	$(COMPOSE_PROD) logs -f

# ----------------------------------------------------------------------
# UTILITIES / DEBUGGING
# ----------------------------------------------------------------------
server-sh: ## [VN] Mở terminal sh trong container Go Server
	$(COMPOSE_DEV) exec server sh

nginx-sh: ## [VN] Mở terminal sh trong container Nginx
	$(COMPOSE_DEV) exec nginx sh

nginx-reload: ## [VN] Reload Nginx config không cần restart container
	$(COMPOSE_DEV) exec nginx nginx -s reload

psql-cli: ## [VN] Mở PostgreSQL CLI (psql) trong container postgres
	$(COMPOSE_DEV) exec postgres psql -U terra_user -d terra_db

db-cli: psql-cli ## [VN] Mở Database CLI (alias cho psql-cli)

clean: ## [VN] Dừng và xóa toàn bộ container, network, volumes của dự án
	$(COMPOSE_DEV) down -v --remove-orphans

# ----------------------------------------------------------------------
# MIGRATION BY GOLANG-MIGRATE V4
# ----------------------------------------------------------------------
DB_MIGRATIONS_DIR = migrations
SERVER_CONTAINER = terra_server_dev

migration: ## [VN] Tạo migration file mới: make migration name=<tên>
	docker exec $(SERVER_CONTAINER) migrate create -ext sql -dir $(DB_MIGRATIONS_DIR) -seq $(name)

migrate-up: ## [VN] Chạy tất cả các migration mới nhất chưa được áp dụng lên database
	docker exec -it $(SERVER_CONTAINER) migrate -path $(DB_MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down: ## [VN] Rollback (hoàn tác) 1 bước migration gần nhất
	docker exec -it $(SERVER_CONTAINER) migrate -path $(DB_MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-reset: ## [VN] Xóa sạch database (drop) và chạy lại toàn bộ migration từ đầu
	docker exec -it $(SERVER_CONTAINER) migrate -path $(DB_MIGRATIONS_DIR) -database "$(DB_URL)" drop
	docker exec -it $(SERVER_CONTAINER) migrate -path $(DB_MIGRATIONS_DIR) -database "$(DB_URL)" up