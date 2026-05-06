.PHONY: dev build docker clean

# Development
dev-frontend:
	cd frontend && npm run dev

dev-backend:
	cd backend && go run ./cmd/server

dev:
	@echo "Run 'make dev-frontend' and 'make dev-backend' in separate terminals"

# Build
build-frontend:
	cd frontend && npm ci && npm run build

build-backend: build-frontend
	mkdir -p backend/cmd/server/frontend/dist
	cp -r frontend/dist/* backend/cmd/server/frontend/dist/
	cd backend && CGO_ENABLED=0 go build -o ../bin/unbound-ui ./cmd/server

build: build-backend

# Docker
docker-build:
	docker compose -f docker/docker-compose.yml build

docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-logs:
	docker compose -f docker/docker-compose.yml logs -f

# Clean
clean:
	rm -rf bin/
	rm -rf frontend/dist
	rm -rf backend/cmd/server/frontend
