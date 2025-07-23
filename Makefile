APP_NAME = gogeturl
APP_CMD_DIR = ./cmd/$(APP_NAME)
BINARY_NAME = $(APP_NAME)
PORT ?= 8080

# Default target
.PHONY: default
default: help

# Build the Go binary
.PHONY: build
build:
	go build -o $(BINARY_NAME) $(APP_CMD_DIR)

# Run the application
.PHONY: run
run:
	PORT=$(PORT) ./$(BINARY_NAME)

# Run tests
.PHONY: test
test:
	go test ./... -v

# Clean built files
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

# Run using Docker
.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run:
	docker run --rm -p $(PORT):$(PORT) $(APP_NAME):latest

# Run linter if you use one like golangci-lint
.PHONY: lint
lint:
	golangci-lint run

# Help
.PHONY: help
help:
	@echo "Makefile commands:"
	@echo "  make build        Build the Go binary"
	@echo "  make run          Run the application locally"
	@echo "  make test         Run tests"
	@echo "  make clean        Remove the built binary"
	@echo "  make docker-build Build Docker image"
	@echo "  make docker-run   Run Docker container"
	@echo "  make lint         Run linters (if configured)"