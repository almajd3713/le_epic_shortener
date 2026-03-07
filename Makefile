
COMPOSE_FILE := infra/docker/docker-compose.yml
COMPOSE_DEV_FILE := infra/docker/docker-compose.dev.yml

.PHONY: up down dev logs restart build help

# Help message to display available commands
help:
	@echo "Available commands:"
	@echo "  up            	- Start the services defined in the docker-compose file"
	@echo "  down          	- Stop the services defined in the docker-compose file"
	@echo "  dev           	- Start the services in development mode (with hot-reloading)"
	@echo "  logs          	- View the logs of the services"
	@echo "  restart       	- Restart the services"
	@echo "  build         	- Build the services (useful if you have made changes to the Dockerfiles)"
	@echo "  dev-build     	- Build and start the services in development mode"
	@echo "  clean         	- Stop and remove all containers, networks, and volumes"
	@echo "  up-deps       	- Spin up dependency services (e.g., database) without starting the main application"
	@echo "  down-deps     	- Stop only the dependency services"
	@echo "  test-unit     	- Run unit tests and generate coverage report"
	@echo "  test-integration 	- Run integration tests (requires the services to be running)"
	@echo "  help          	- Display this help message"

# Start the services defined in the docker-compose file
up:
	docker compose -f $(COMPOSE_FILE) up -d

# Stop the services defined in the docker-compose file
down:
	docker compose -f $(COMPOSE_FILE) down

# Start the services in development mode (with hot-reloading)
dev:
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_DEV_FILE) up -d

# View the logs of the services
logs:
	docker compose -f $(COMPOSE_FILE) logs -f

# Restart the services
restart:
	docker compose -f $(COMPOSE_FILE) restart

# Build the services (useful if you have made changes to the Dockerfiles)
build:
	docker compose -f $(COMPOSE_FILE) build

# Build and start the services in development mode
dev-build:
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_DEV_FILE) build
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_DEV_FILE) up -d

# Stop and remove all containers, networks, and volumes
clean:
	docker compose -f $(COMPOSE_FILE) down -v --remove-orphans
	docker images --filter=reference="shortener*" -q | xargs -r docker rmi -f

# Spin up dependency services (e.g., database) without starting the main application
up-deps:
	docker compose -f $(COMPOSE_FILE) up -d db redis

# Stop only the dependency services
down-deps:
	docker compose -f $(COMPOSE_FILE) stop db redis

# Run unit tests
test-unit:
	cd backend && go tool cover -html=coverage.unit.out -o coverage.html
	@echo "Coverage report generated: backend/coverage.html"

# Integration tests (requires the services to be running)
test-integration: up-deps
	cd backend && go test -v -race -tags=integration ./...; \
	$(MAKE) down-deps

