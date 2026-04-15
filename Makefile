.PHONY: proto build up down

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

down:
	docker-compose down -v
	@echo "Shutting down"