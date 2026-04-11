# Переменные для удобства
DOCKER_COMPOSE = docker-compose -f deployments/docker-compose.yaml
PROTO_DIR = pkg/proto
MODULE_NAME = github.com/max/camp

.PHONY: proto build up down


proto:
	protoc --proto_path=. \
           --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           api/proto/admin.proto
	@echo "code generated"

build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) up -d
	@echo "Services are running"

down:
	$(DOCKER_COMPOSE) down
