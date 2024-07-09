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
	docker build -f ./build/demo/Dockerfile -t zaytcevcom/go-msa:1.0.7 .
	docker build -f ./build/migrations/Dockerfile -t zaytcevcom/go-msa-migrations:1.0.7 .
	docker build -f ./build/auth/Dockerfile -t zaytcevcom/go-msa-auth:1.0.7 .
	docker build -f ./build/order/Dockerfile -t zaytcevcom/go-msa-order:1.0.7 .
	docker build -f ./build/billing/Dockerfile -t zaytcevcom/go-msa-billing:1.0.7 .
	docker build -f ./build/account_creator/Dockerfile -t zaytcevcom/go-msa-billing-account-creator:1.0.7 .
	docker build -f ./build/notification/Dockerfile -t zaytcevcom/go-msa-notification:1.0.7 .
	docker build -f ./build/notification_sender/Dockerfile -t zaytcevcom/go-msa-notification-sender:1.0.7 .

dockerhub-push:
	docker push zaytcevcom/go-msa:1.0.7
	docker push zaytcevcom/go-msa-migrations:1.0.7
	docker push zaytcevcom/go-msa-auth:1.0.7
	docker push zaytcevcom/go-msa-order:1.0.7
	docker push zaytcevcom/go-msa-billing:1.0.7
	docker push zaytcevcom/go-msa-billing-account-creator:1.0.7
	docker push zaytcevcom/go-msa-notification:1.0.7
	docker push zaytcevcom/go-msa-notification-sender:1.0.7

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
	helm-prometheus helm-postgres \
	k8s-rabbitmq \
	k8s-redis \
	helm-demo \
	helm-auth \
	helm-order \
	helm-billing helm-billing-account-creator \
	helm-notification helm-notification-sender \
	k8s-ingress

helm-prometheus:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts && \
	helm repo update && \
	helm install stack prometheus-community/kube-prometheus-stack -f ./deployments/helm/prometheus/values.yaml

helm-postgres:
	helm repo add bitnami https://charts.bitnami.com/bitnami && \
	helm repo update && \
	helm install db bitnami/postgresql -f ./deployments/helm/postgresql/values.yaml

helm-demo:
	kubectl create configmap demo-config --from-file=configs/demo/config.json && \
	helm install demo ./deployments/helm/demo

helm-auth:
	kubectl create configmap auth-config --from-file=configs/auth/config.json && \
	helm install auth ./deployments/helm/auth

helm-order:
	kubectl create configmap order-config --from-file=configs/order/config.json && \
	helm install order ./deployments/helm/order

helm-billing:
	kubectl create configmap billing-config --from-file=configs/billing/config.json && \
	helm install billing ./deployments/helm/billing

helm-billing-account-creator:
	kubectl create configmap account-creator-config --from-file=configs/account_creator/config.json && \
	helm install account-creator ./deployments/helm/account_creator

helm-notification:
	kubectl create configmap notification-config --from-file=configs/notification/config.json && \
	helm install notification ./deployments/helm/notification

helm-notification-sender:
	kubectl create configmap notification-sender-config --from-file=configs/notification_sender/config.json && \
	helm install notification-sender ./deployments/helm/notification_sender

k8s: k8s-init \
	k8s-demo \
	k8s-ingress

k8s-init:
	minikube delete && \
    minikube start

k8s-demo:
	kubectl create configmap demo-config --from-file=configs/demo/config.json && \
	kubectl apply -f ./deployments/k8s/demo

k8s-rabbitmq:
	kubectl apply -f ./deployments/k8s/rabbitmq

k8s-redis:
	kubectl apply -f ./deployments/k8s/redis

k8s-ingress:
	kubectl apply -f ./deployments/k8s/ingress && \
	minikube addons enable ingress && \
	minikube tunnel


clear:
	rm -rf var/postgres/*

wait-postgres:
	sleep 10

migrations-migrate:
	goose -dir migrations postgres "user=test password=test dbname=demo host=localhost sslmode=disable" up

ab:
	@bash -c 'for i in {1..25}; do ab -c 250 -n 1000 "http://arch.homework/user/1"; done'
