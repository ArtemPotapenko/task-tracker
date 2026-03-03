.PHONY: tools proto proto-external proto-internal gateway openapi clean-proto

GOPATH := $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

PROTOC ?= protoc

EXTERNAL_PROTO_DIR := api/proto/external
INTERNAL_PROTO_DIR := api/proto/internal
GOOGLEAPIS_DIR := third_party/googleapis

GEN_EXTERNAL_DIR := gen/public
GEN_INTERNAL_DIR := gen/private

PROTO_EXTERNAL_FILES := \
	$(EXTERNAL_PROTO_DIR)/account/auth.proto \
	$(EXTERNAL_PROTO_DIR)/task/task.proto

PROTO_INTERNAL_FILES := \
	$(INTERNAL_PROTO_DIR)/account/users.proto \
	$(INTERNAL_PROTO_DIR)/scheduler/scheduler.proto

tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.22.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.22.0

proto: proto-external proto-internal

proto-external:
	mkdir -p gen/public/openapi
	$(PROTOC) -I $(EXTERNAL_PROTO_DIR) -I $(GOOGLEAPIS_DIR) \
		--go_out=$(GEN_EXTERNAL_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_EXTERNAL_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_EXTERNAL_DIR) --grpc-gateway_opt=logtostderr=true,paths=source_relative \
		--openapiv2_out=gen/public/openapi --openapiv2_opt=logtostderr=true \
		$(PROTO_EXTERNAL_FILES)

proto-internal:
	$(PROTOC) -I $(INTERNAL_PROTO_DIR) -I $(GOOGLEAPIS_DIR) \
		--go_out=$(GEN_INTERNAL_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_INTERNAL_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_INTERNAL_FILES)

gateway: proto-external

openapi: proto-external

clean-proto:
	rm -rf $(GEN_EXTERNAL_DIR)/account $(GEN_EXTERNAL_DIR)/task gen/public/openapi
	rm -rf $(GEN_INTERNAL_DIR)/account $(GEN_INTERNAL_DIR)/scheduler
