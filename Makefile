.PHONY: dev dev-api dev-web dev-admin test lint migrate-up migrate-down migrate-status sqlc build docker-up docker-down

GOPATH ?= $(shell go env GOPATH)
GOOSE := $(GOPATH)/bin/goose
SQLC := $(GOPATH)/bin/sqlc

# MySQL credentials (default matches local MySQL root/GilangRizky)
DB_USER ?= root
DB_PASS ?= GilangRizky
DB_HOST ?= 127.0.0.1
DB_PORT ?= 3306
DB_NAME ?= news_portal
DB_DSN ?= "$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?parseTime=true"

dev-infra:
	@echo "==> Starting local MySQL & Redis containers..."
	@docker start mysql-dev 2>/dev/null || true
	@docker start redis-dev 2>/dev/null || true
	@echo "==> MySQL & Redis containers ready."

clean-data:
	@echo "==> Cleaning all dummy articles, views, revisions, tags, redirects, and media..."
	@mysql -u $(DB_USER) -p$(DB_PASS) -h $(DB_HOST) $(DB_NAME) -e "\
		SET FOREIGN_KEY_CHECKS = 0; \
		TRUNCATE TABLE article_revisions; \
		TRUNCATE TABLE article_tag; \
		TRUNCATE TABLE article_views_daily; \
		TRUNCATE TABLE articles; \
		TRUNCATE TABLE tags; \
		TRUNCATE TABLE redirects; \
		TRUNCATE TABLE audit_logs; \
		TRUNCATE TABLE media; \
		DELETE FROM categories WHERE id > 6; \
		DELETE FROM authors WHERE id > 3; \
		DELETE FROM pages WHERE id > 4; \
		DELETE FROM users WHERE id > 3; \
		SET FOREIGN_KEY_CHECKS = 1;"
	@docker exec redis-dev redis-cli FLUSHALL 2>/dev/null || true
	@echo "==> Database and Redis cache are now clean like brand new."

dev:
	@echo "==> Starting local infrastructure & services..."
	@docker start mysql-dev 2>/dev/null || true
	@docker start redis-dev 2>/dev/null || true
	@echo "==> MySQL & Redis are ready."
	@echo "==> Running API in development mode..."
	cd apps/api && go run ./cmd/server/main.go

dev-api:
	cd apps/api && go run ./cmd/server/main.go

dev-web:
	cd apps/web && npm run dev

dev-admin:
	cd apps/admin && npm run dev

test:
	@echo "==> Running Go backend tests..."
	cd apps/api && go test -v -race ./...

lint:
	@echo "==> Checking Go backend code formatting & vet..."
	cd apps/api && go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd apps/api && golangci-lint run; \
	fi

migrate-up:
	@echo "==> Running database migrations UP..."
	cd apps/api/db && $(GOOSE) -dir migrations mysql $(DB_DSN) up

migrate-down:
	@echo "==> Rolling back 1 database migration..."
	cd apps/api/db && $(GOOSE) -dir migrations mysql $(DB_DSN) down

migrate-status:
	@echo "==> Checking migration status..."
	cd apps/api/db && $(GOOSE) -dir migrations mysql $(DB_DSN) status

seed:
	@echo "==> Running database seeder (>= 1,000 articles)..."
	cd apps/api && go run ./cmd/seed/main.go

sqlc:
	@echo "==> Generating type-safe SQL queries with sqlc..."
	cd apps/api/db && $(SQLC) generate

build:
	@echo "==> Building Go API binary..."
	cd apps/api && CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/server ./cmd/server/main.go
	@echo "==> Build complete."

docker-up:
	docker compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker compose -f deploy/docker-compose.yml down
