.PHONY: proto build up down downv gendocs

PROTO_DIR=./api/proto
OUT_DIR=./api/proto/proto_output
proto:
	@mkdir -p $(OUT_DIR)
	protoc \
    		-I=$(PROTO_DIR) \
    		--go_out=$(OUT_DIR) \
    		--go_opt=paths=source_relative \
    		--go-grpc_out=$(OUT_DIR) \
    		--go-grpc_opt=paths=source_relative \
    		$(shell find $(PROTO_DIR) -name "*.proto")
	@echo "code generated"

build:
	docker-compose up --build -d
	@echo "Services are running"

up:
	docker-compose up -v
	@echo "Services are running"

downv:
	docker-compose down -v
	@echo "Shutting down with clear volumes"
down:
	docker-compose down
	@echo "shutdown"
gendocs:
	swag init -g cmd/gateway/main.go -o cmd/gateway/docs
	@echo "Documentation was generated successfully"