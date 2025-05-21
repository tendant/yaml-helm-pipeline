.PHONY: build run clean dev docker-build docker-run docker-compose-up docker-compose-down

# Build the backend
build:
	go build -o bin/yaml-helm-pipeline ./cmd/server/

# Run the backend
run: build
	./bin/yaml-helm-pipeline

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/dist/

docker-push:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--push \
		--tag wang/yaml-helm-pipeline:$(VERSION) \
		--tag wang/yaml-helm-pipeline:latest \
		.


# Run in development mode
dev:
	./dev.sh

# Build frontend
frontend-build:
	cd frontend && npm install && npm run build

# Build Docker image
docker-build:
	docker build -t yaml-helm-pipeline .

# Run Docker container
docker-run:
	docker run -p 4000:4000 \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e REPO_OWNER=$(REPO_OWNER) \
		-e REPO_NAME=$(REPO_NAME) \
		yaml-helm-pipeline

# Build and run in Docker
docker: docker-build docker-run

# Docker Compose up
docker-compose-up:
	docker-compose up -d

# Docker Compose down
docker-compose-down:
	docker-compose down

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the backend"
	@echo "  run           - Run the backend"
	@echo "  clean         - Clean build artifacts"
	@echo "  dev           - Run in development mode"
	@echo "  frontend-build - Build the frontend"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker        - Build and run in Docker"
	@echo "  docker-compose-up   - Start with Docker Compose"
	@echo "  docker-compose-down - Stop Docker Compose containers"
	@echo "  help          - Show this help message"
