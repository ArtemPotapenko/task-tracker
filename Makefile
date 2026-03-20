.PHONY: tools proto proto-external proto-internal gateway openapi clean-proto migrate-account-up migrate-account-down migrate-task-up migrate-task-down test-unit test-integration test-e2e build lint

GOPATH := $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

PROTOC ?= protoc
GOOSE ?= goose
GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOLANGCI_LINT_VERSION ?= v1.64.8

ACCOUNT_DB_DSN ?= postgres://admin:secret@localhost:5433/accountdb?sslmode=disable
TASK_DB_DSN ?= postgres://admin:secret@localhost:5434/taskdb?sslmode=disable
GOCACHE ?= /tmp/go-build-cache

ACCOUNT_MIGRATIONS_DIR := migrations/account
TASK_MIGRATIONS_DIR := migrations/task
PROTO_INCLUDE_DIR := api/proto/include

UNIT_TEST_PKGS := ./internal/account/usecase ./internal/task/usecase ./internal/email/usecase ./internal/email/transport/kafka
INTEGRATION_TEST_PKGS := ./internal/account/repo ./internal/task/repo
E2E_TEST_PKGS := ./internal/account/app ./internal/task/app ./internal/scheduler/app

EXTERNAL_PROTO_DIR := api/proto/external
INTERNAL_PROTO_DIR := api/proto/internal
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
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

proto: proto-external proto-internal

proto-external:
	mkdir -p gen/public/openapi
	$(PROTOC) -I $(EXTERNAL_PROTO_DIR) -I $(PROTO_INCLUDE_DIR) \
		--go_out=$(GEN_EXTERNAL_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_EXTERNAL_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_EXTERNAL_DIR) --grpc-gateway_opt=logtostderr=true,paths=source_relative \
		--openapiv2_out=gen/public/openapi --openapiv2_opt=logtostderr=true \
		$(PROTO_EXTERNAL_FILES)

proto-internal:
	$(PROTOC) -I $(INTERNAL_PROTO_DIR) -I $(PROTO_INCLUDE_DIR) \
		--go_out=$(GEN_INTERNAL_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_INTERNAL_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_INTERNAL_FILES)

gateway: proto-external

openapi: proto-external

clean-proto:
	rm -rf $(GEN_EXTERNAL_DIR)/account $(GEN_EXTERNAL_DIR)/task gen/public/openapi
	rm -rf $(GEN_INTERNAL_DIR)/account $(GEN_INTERNAL_DIR)/scheduler

migrate-account-up:
	$(GOOSE) -dir $(ACCOUNT_MIGRATIONS_DIR) postgres "$(ACCOUNT_DB_DSN)" up

migrate-account-down:
	$(GOOSE) -dir $(ACCOUNT_MIGRATIONS_DIR) postgres "$(ACCOUNT_DB_DSN)" down

migrate-task-up:
	$(GOOSE) -dir $(TASK_MIGRATIONS_DIR) postgres "$(TASK_DB_DSN)" up

migrate-task-down:
	$(GOOSE) -dir $(TASK_MIGRATIONS_DIR) postgres "$(TASK_DB_DSN)" down

test-unit:
	env GOCACHE=$(GOCACHE) $(GO) test $(UNIT_TEST_PKGS)

test-integration:
	env GOCACHE=$(GOCACHE) TESTCONTAINERS_RYUK_DISABLED=true $(GO) test -count=1 -v $(INTEGRATION_TEST_PKGS)

test-e2e:
	env GOCACHE=$(GOCACHE) $(GO) test $(E2E_TEST_PKGS)

build:
	env GOCACHE=$(GOCACHE) $(GO) build ./...

lint:
	env GOCACHE=$(GOCACHE) $(GOLANGCI_LINT) run ./...
