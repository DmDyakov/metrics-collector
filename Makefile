MODULE := $(shell head -1 go.mod | awk '{print $$2}')

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DATE := $(shell date -u +'%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BIN_DIR := bin

COVERAGE_FILE := coverage.out
COVERAGE_MIN := 50
COVERAGE_EXCLUDE := \
	/cmd/ \
	/mocks \
	/dto \
	/model \
	/logger \
	/pprof \
	/pool \
	/buildinfo \
	/templates

LDFLAGS := -X $(MODULE)/pkg/buildinfo.Version=$(VERSION) \
           -X $(MODULE)/pkg/buildinfo.Date=$(DATE) \
           -X $(MODULE)/pkg/buildinfo.Commit=$(COMMIT)

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: test-short
test-short:
	go test -race -short ./...

.PHONY: test-verbose
test-verbose:
	go test -v -json ./... 2>&1 | gotestfmt

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=$(COVERAGE_FILE) -covermode=atomic \
		$$(go list ./... | grep -v -E '$(subst $(eval) ,|,$(strip $(COVERAGE_EXCLUDE)))')
	go tool cover -func=$(COVERAGE_FILE)

.PHONY: build-server
build-server:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/server ./cmd/server/

.PHONY: build-agent
build-agent:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/agent ./cmd/agent/

.PHONY: build-all
build-all: build-server build-agent

.PHONY: run-server
run-server:
	go run -ldflags "$(LDFLAGS)" ./cmd/server/

.PHONY: run-agent
run-agent:
	go run -ldflags "$(LDFLAGS)" ./cmd/agent/

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -s -w .
	goimports -w .

.PHONY: check-fmt
check-fmt:
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@test -z "$$(goimports -l .)" || (echo "Imports are not formatted. Run 'make fmt'" && exit 1)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(COVERAGE_FILE)

.PHONY: all
all: fmt lint test-coverage build-all