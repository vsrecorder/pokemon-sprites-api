SHELL := /bin/bash

.PHONY: run
run:
	go mod tidy
	go run cmd/main.go

.PHONY: build
build:
	go mod tidy
	go build -o bin/pokemon-sprites-api cmd/main.go

.PHONY: image
image:
	docker build . --no-cache -t vsrecorder/pokemon-sprites-api:local && \
	docker push vsrecorder/pokemon-sprites-api:local

.PHONY: deploy
deploy:
	docker compose pull && docker compose down && docker compose up -d

.PHONY: restart
restart:
	docker compose down && docker compose up -d

.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down

.PHONY: log
log:
	docker logs -f pokemon-sprites-api-pokemon-sprites-api-1
