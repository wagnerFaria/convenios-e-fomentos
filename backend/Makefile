.PHONY: run dev build test test-unit cover-authz lint sqlc migrate-up migrate-down migrate-create seed cleanup docker-up docker-down

GOBIN := $(shell go env GOPATH)/bin
export PATH := $(PATH):$(GOBIN)

run:
	go run ./cmd/api

dev:
	air -c .air.toml

build:
	go build -o bin/api ./cmd/api
	go build -o bin/cli ./cmd/cli

test:
	go test -race -count=1 ./...

test-unit:
	go test -short -race -count=1 ./...

# Falha se a cobertura do pacote authz (testes unitários + integração) ficar abaixo de 80%.
cover-authz:
	go test -race -count=1 -coverprofile=/tmp/authz.cover ./internal/authz
	@go tool cover -func=/tmp/authz.cover | awk '/^total:/ { gsub("%","",$$3); print "cobertura do authz: " $$3 "%"; if ($$3+0 < 80) { print "abaixo de 80%"; exit 1 } }'

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

migrate-up:
	go run ./cmd/cli migrate:up

migrate-down:
	go run ./cmd/cli migrate:down

# uso: make migrate-create name=criar_tabela_x
migrate-create:
	goose -dir internal/database/migrations create $(name) sql

seed:
	go run ./cmd/cli seed

cleanup:
	go run ./cmd/cli cleanup

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
