ifeq ($(OS),Windows_NT)
    ARCH = windows
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
			ARCH = linux
    endif
    ifeq ($(UNAME_S),Darwin)
			ARCH = darwin
    endif
endif
REPO_VERSION := $$(git describe --abbrev=0 --tags)
BUILD_DATE := $$(date +%Y-%m-%d-%H:%M)
GIT_HASH := $$(git rev-parse --short HEAD)
GOBUILD_VERSION_ARGS := -ldflags "-s -X main.Version=$(REPO_VERSION) -X main.GitCommit=$(GIT_HASH) -X main.BuildDate=$(BUILD_DATE)"
BINARY := cddamap
MAIN_PKG := github.com/ralreegorganon/cddamap/cmd/cddamap

REGISTRY := registry.ralreegorganon.com
IMAGE_NAME := $(BINARY)

DB_USER := cddamap
DB_PASSWORD := cddamap
DB_PORT_MIGRATION := 9432

CDDAMAP_CONNECTION_STRING_LOCAL := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_USER)?sslmode=disable
CDDAMAP_CONNECTION_STRING_DOCKER := postgres://$(DB_USER):$(DB_PASSWORD)@db:5432/$(DB_USER)?sslmode=disable
CDDAMAP_CONNECTION_STRING_MIGRATION_DOCKER := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT_MIGRATION)/$(DB_USER)?sslmode=disable

CDDAMAP_MIGRATIONS_PATH := file://internal/server/migrations

build:
	go build -i -v -o images/$(BINARY)/bin/$(ARCH)/$(BINARY) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

run: build
	CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_LOCAL)" CDDAMAP_MIGRATIONS_PATH="$(CDDAMAP_MIGRATIONS_PATH)"  ./images/$(BINARY)/bin/$(ARCH)/$(BINARY)

install:
	go install $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)

migrate:
	cd internal/server/migrations/ && CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_LOCAL)" ./run-migrations

docker:
	mkdir -p images/$(BINARY)/migrations && cp internal/server/migrations/*.sql images/$(BINARY)/migrations
	GOOS=linux GOARCH=amd64 go build -o images/$(BINARY)/bin/linux/$(BINARY) $(GOBUILD_VERSION_ARGS) $(MAIN_PKG)
	docker build --pull -t $(REGISTRY)/$(IMAGE_NAME):latest images/$(BINARY)

run-docker: docker
	cd images/$(BINARY)/ && DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_DOCKER)" docker-compose -p $(BINARY) rm -f $(BINARY)
	DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_DOCKER)" docker-compose -f images/$(BINARY)/docker-compose.yml -p $(BINARY) build
	DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_DOCKER)" docker-compose -f images/$(BINARY)/docker-compose.yml -p $(BINARY) up -d

stop-docker:
	cd images/$(BINARY)/ && DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_DOCKER)" docker-compose -p $(BINARY) stop

migrate-docker:
	cd internal/server/migrations/ && CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_MIGRATION_DOCKER)" ./run-migrations

docker-logs: 
	cd images/$(BINARY)/ && DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_PORT_MIGRATION=$(DB_PORT_MIGRATION) CDDAMAP_CONNECTION_STRING="$(CDDAMAP_CONNECTION_STRING_DOCKER)" docker-compose -p $(BINARY) logs

clean:
	rm -rf images/$(BINARY)/bin/*
	rm -rf images/$(BINARY)/migrations/*

release: docker
	docker push $(REGISTRY)/$(IMAGE_NAME):latest
	docker tag $(REGISTRY)/$(IMAGE_NAME):latest $(REGISTRY)/$(IMAGE_NAME):$(REPO_VERSION)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(REPO_VERSION)

.PHONY: build install
