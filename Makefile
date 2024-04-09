BIN := "./bin/demo"
DOCKER_IMG="demo:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

init: docker-down-clear \
	docker-network \
	docker-pull docker-build docker-up

up: docker-network docker-up
down: docker-down
restart: down up

docker-pull:
	docker compose -f ./deployments/development/docker-compose.yml pull

docker-build:
	docker compose -f ./deployments/development/docker-compose.yml build --pull

dockerhub-build-amd64:
	docker build --platform linux/amd64 -f ./build/demo/Dockerfile -t zaytcevcom/go-msa:1.0.1 .

dockerhub-push:
	docker push zaytcevcom/go-msa:1.0.1

docker-up:
	docker compose -f ./deployments/development/docker-compose.yml up -d

docker-down:
	docker compose -f ./deployments/development/docker-compose.yml down --remove-orphans

docker-down-clear:
	docker compose -f ./deployments/development/docker-compose.yml down -v --remove-orphans

docker-network:
	docker network create demo_network || true

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/demo

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version


test:
	go test -race -count 100 ./internal/...


remove-lint-deps:
	rm $(which golangci-lint)

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint