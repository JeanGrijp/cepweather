.PHONY: run test build docker-build docker-run compose clean docker-watch

WEATHER_API_KEY ?=
IMAGE ?= cepweather
BIN ?= bin/server
GOCACHE_DIR := $(CURDIR)/.cache

run:
	@if [ -z "$(WEATHER_API_KEY)" ]; then \
		echo "WEATHER_API_KEY environment variable is required"; \
		exit 1; \
	fi
	WEATHER_API_KEY=$(WEATHER_API_KEY) PORT=8080 go run ./cmd/api

test:
	GOCACHE=$(GOCACHE_DIR) go test ./...

build:
	@mkdir -p $(dir $(BIN))
	GOCACHE=$(GOCACHE_DIR) CGO_ENABLED=0 go build -o $(BIN) ./cmd/api

docker-build:
	docker build -t $(IMAGE) .

docker-run: docker-build
	@if [ -z "$(WEATHER_API_KEY)" ]; then \
		echo "WEATHER_API_KEY environment variable is required"; \
		exit 1; \
	fi
	docker run --rm -p 8080:8080 -e WEATHER_API_KEY=$(WEATHER_API_KEY) $(IMAGE)

compose:
	@if [ -z "$(WEATHER_API_KEY)" ]; then \
		echo "WEATHER_API_KEY environment variable is required"; \
		exit 1; \
	fi
	WEATHER_API_KEY=$(WEATHER_API_KEY) docker compose up --build

clean:
	rm -rf $(BIN) $(GOCACHE_DIR)

docker-watch:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && docker compose up --build; \
	else \
		if [ -z "$(WEATHER_API_KEY)" ]; then \
			echo "WEATHER_API_KEY environment variable is required or create a .env file"; \
			exit 1; \
		fi; \
		docker compose up --build; \
	fi