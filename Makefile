BIN := "./bin/demo"
DOCKER_IMG="demo:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

init: docker-down-clear \
	clear \
	docker-network \
	docker-pull docker-build docker-up \
	wait-postgres migrations-migrate

up: docker-network docker-up wait-postgres migrations-migrate
down: docker-down
restart: down up

docker-pull:
	docker compose -f ./deployments/development/docker-compose.yml pull

docker-build:
	docker compose -f ./deployments/development/docker-compose.yml build --pull

dockerhub-build-amd64:
	docker build -f ./build/demo/Dockerfile -t zaytcevcom/go-msa:1.0.3 .
	docker build -f ./build/migrations/Dockerfile -t zaytcevcom/go-msa-migrations:1.0.1 .

dockerhub-push:
	docker push zaytcevcom/go-msa:1.0.3
	docker push zaytcevcom/go-msa-migrations:1.0.1

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


helm: k8s-init \
	helm-postgres helm-demo \
	k8s-ingress

helm-postgres:
	helm repo add bitnami https://charts.bitnami.com/bitnami && \
	helm install db bitnami/postgresql -f ./deployments/helm/postgresql/values.yaml

helm-demo:
	kubectl create configmap demo-config --from-file=configs/demo/config.json && \
	helm install demo ./deployments/helm/demo

k8s: k8s-init \
	k8s-demo \
	k8s-ingress

k8s-init:
	minikube delete && \
    minikube start

k8s-demo:
	kubectl create configmap demo-config --from-file=configs/demo/config.json && \
	kubectl apply -f ./deployments/k8s/demo

k8s-ingress:
	kubectl apply -f ./deployments/k8s/ && \
	minikube addons enable ingress && \
	minikube tunnel


clear:
	rm -rf var/postgres/*

wait-postgres:
	sleep 10

migrations-migrate:
	goose -dir migrations postgres "user=test password=test dbname=demo host=localhost sslmode=disable" up