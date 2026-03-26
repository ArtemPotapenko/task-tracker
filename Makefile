.PHONY: tools tidy proto proto-external proto-internal clean-proto migrate-account-up migrate-account-down migrate-task-up migrate-task-down test test-unit test-integration test-e2e build lint

GOPATH := $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

GO ?= go
GOOSE ?= goose
GOLANGCI_LINT ?= golangci-lint
GOLANGCI_LINT_VERSION ?= v1.64.8
PGV_VERSION ?= v1.0.2
GOCACHE ?= /tmp/go-build-cache
MODULE_DIRS := shared-libs proto-lib account-service task-service email-service gateway scheduler-service

ACCOUNT_DB_DSN ?= postgres://admin:secret@localhost:5433/accountdb?sslmode=disable
TASK_DB_DSN ?= postgres://admin:secret@localhost:5434/taskdb?sslmode=disable
PROTOC ?= protoc
WORKSPACE_PKGS := ./account-service/... ./task-service/... ./email-service/... ./gateway/... ./scheduler-service/... ./shared-libs/... ./proto-lib/...
UNIT_TEST_PKGS := ./account-service/internal/usecase ./task-service/internal/usecase ./email-service/internal/usecase ./email-service/internal/transport/kafka
INTEGRATION_TEST_PKGS := ./account-service/internal/repo ./task-service/internal/repo
E2E_TEST_PKGS := ./account-service/internal/app ./task-service/internal/app ./scheduler-service/internal/app

PROTO_ROOT_DIR := proto-lib
PROTO_INCLUDE_DIR := $(PROTO_ROOT_DIR)/api/proto/include
PGV_INCLUDE_DIR := $(shell $(GO) env GOMODCACHE)/github.com/envoyproxy/protoc-gen-validate@$(PGV_VERSION)
EXTERNAL_PROTO_DIR := $(PROTO_ROOT_DIR)/api/proto/external
INTERNAL_PROTO_DIR := $(PROTO_ROOT_DIR)/api/proto/internal
GEN_EXTERNAL_DIR := $(PROTO_ROOT_DIR)/gen/public
GEN_INTERNAL_DIR := $(PROTO_ROOT_DIR)/gen/private

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
	go install github.com/envoyproxy/protoc-gen-validate@$(PGV_VERSION)
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

tidy:
	@for dir in $(MODULE_DIRS); do \
		echo "==> go mod tidy $$dir"; \
		( cd $$dir && $(GO) mod tidy ); \
	done

proto: proto-external proto-internal

proto-external:
	mkdir -p $(GEN_EXTERNAL_DIR)/openapi
	$(PROTOC) -I $(EXTERNAL_PROTO_DIR) -I $(PROTO_INCLUDE_DIR) -I $(PGV_INCLUDE_DIR) \
		--go_out=$(GEN_EXTERNAL_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_EXTERNAL_DIR) --go-grpc_opt=paths=source_relative \
		--validate_out=$(GEN_EXTERNAL_DIR) --validate_opt=lang=go,paths=source_relative \
		--grpc-gateway_out=$(GEN_EXTERNAL_DIR) --grpc-gateway_opt=logtostderr=true,paths=source_relative \
		--openapiv2_out=$(GEN_EXTERNAL_DIR)/openapi --openapiv2_opt=logtostderr=true \
		$(PROTO_EXTERNAL_FILES)

proto-internal:
	$(PROTOC) -I $(INTERNAL_PROTO_DIR) -I $(PROTO_INCLUDE_DIR) -I $(PGV_INCLUDE_DIR) \
		--go_out=$(GEN_INTERNAL_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_INTERNAL_DIR) --go-grpc_opt=paths=source_relative \
		--validate_out=$(GEN_INTERNAL_DIR) --validate_opt=lang=go,paths=source_relative \
		$(PROTO_INTERNAL_FILES)

clean-proto:
	rm -rf $(GEN_EXTERNAL_DIR)/account $(GEN_EXTERNAL_DIR)/task $(GEN_EXTERNAL_DIR)/openapi
	rm -rf $(GEN_INTERNAL_DIR)/account $(GEN_INTERNAL_DIR)/scheduler

migrate-account-up:
	$(GOOSE) -dir account-service/migrations postgres "$(ACCOUNT_DB_DSN)" up

migrate-account-down:
	$(GOOSE) -dir account-service/migrations postgres "$(ACCOUNT_DB_DSN)" down

migrate-task-up:
	$(GOOSE) -dir task-service/migrations postgres "$(TASK_DB_DSN)" up

migrate-task-down:
	$(GOOSE) -dir task-service/migrations postgres "$(TASK_DB_DSN)" down

test:
	env GOCACHE=$(GOCACHE) $(GO) test $(WORKSPACE_PKGS)

test-unit:
	env GOCACHE=$(GOCACHE) $(GO) test $(UNIT_TEST_PKGS)

test-integration:
	env GOCACHE=$(GOCACHE) TESTCONTAINERS_RYUK_DISABLED=true $(GO) test -count=1 -v $(INTEGRATION_TEST_PKGS)

test-e2e:
	env GOCACHE=$(GOCACHE) $(GO) test $(E2E_TEST_PKGS)

build:
	env GOCACHE=$(GOCACHE) $(GO) build $(WORKSPACE_PKGS)

lint:
	env GOCACHE=$(GOCACHE) $(GOLANGCI_LINT) run $(WORKSPACE_PKGS)
