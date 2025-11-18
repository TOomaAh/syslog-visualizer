.PHONY: help build up down restart logs ps clean test backend-build frontend-build

# Default target
help:
	@echo "Syslog Visualizer - Makefile Commands"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make up              - Start all services"
	@echo "  make down            - Stop all services"
	@echo "  make restart         - Restart all services"
	@echo "  make build           - Build all images"
	@echo "  make rebuild         - Rebuild and restart all services"
	@echo "  make logs            - View logs (all services)"
	@echo "  make logs-backend    - View backend logs"
	@echo "  make logs-frontend   - View frontend logs"
	@echo "  make ps              - Show running containers"
	@echo "  make clean           - Stop and remove containers, networks, volumes"
	@echo ""
	@echo "Individual Services:"
	@echo "  make backend         - Start only backend"
	@echo "  make frontend        - Start only frontend"
	@echo "  make backend-build   - Build backend image"
	@echo "  make frontend-build  - Build frontend image"
	@echo ""
	@echo "Development:"
	@echo "  make dev-backend     - Run backend locally (go run)"
	@echo "  make dev-frontend    - Run frontend locally (npm run dev)"
	@echo "  make test            - Run tests"
	@echo ""
	@echo "Utilities:"
	@echo "  make shell-backend   - Shell into backend container"
	@echo "  make shell-frontend  - Shell into frontend container"
	@echo "  make backup          - Backup database volume"
	@echo "  make stats           - Show container stats"

# Docker Compose commands
up:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose restart

build:
	docker-compose build

rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

logs:
	docker-compose logs -f

logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

ps:
	docker-compose ps

clean:
	docker-compose down -v
	docker system prune -f

# Individual services
backend:
	docker-compose up -d backend

frontend:
	docker-compose up -d frontend

backend-build:
	docker-compose build backend

frontend-build:
	docker-compose build frontend

# Development (local)
dev-backend:
	go run cmd/server/main.go

dev-frontend:
	cd web && npm run dev

# Testing
test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Utilities
shell-backend:
	docker-compose exec backend sh

shell-frontend:
	docker-compose exec frontend sh

backup:
	docker run --rm \
		-v syslog-visualizer_syslog-data:/data \
		-v $(PWD):/backup \
		alpine tar czf /backup/syslog-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz /data
	@echo "Backup created: syslog-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz"

stats:
	docker stats

# Health checks
health:
	@echo "Checking backend health..."
	@curl -s http://localhost:8080/api/health | jq . || echo "Backend not responding"
	@echo ""
	@echo "Checking frontend health..."
	@curl -s http://localhost:3000 > /dev/null && echo "Frontend OK" || echo "Frontend not responding"

# Production
prod-up:
	docker-compose -f docker-compose.prod.yml up -d

prod-down:
	docker-compose -f docker-compose.prod.yml down

prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

# Install development dependencies
install:
	go mod download
	cd web && npm install

# Format code
fmt:
	go fmt ./...
	cd web && npm run lint

# Build for production
build-prod:
	go build -o bin/syslog-server cmd/server/main.go
	cd web && npm run build
