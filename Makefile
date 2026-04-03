.PHONY: build test test-unit test-race lint vet fmt fmt-check tidy tidy-check \
	check release-check release clean help tools

BINARY := $(CURDIR)/bin/hush
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.DEFAULT_GOAL := check

help:
	@echo "hush — Dotfiles for AI Agent Rules"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build the CLI"
	@echo "  make test        Run all tests"
	@echo "  make test-unit   Run unit tests"
	@echo "  make test-race   Run tests with race detector"
	@echo "  make lint        Run golangci-lint"
	@echo "  make vet         Run go vet"
	@echo "  make fmt         Format Go source"
	@echo "  make fmt-check   Check formatting (CI gate)"
	@echo "  make tidy        Tidy dependencies"
	@echo "  make tidy-check  Verify go.mod/go.sum tidiness"
	@echo "  make check         Full CI gate (fmt + vet + test)"
	@echo "  make release-check Pre-release validation"
	@echo "  make release       Tag and push a release"
	@echo "  make clean         Remove build artifacts"

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/hush

test: test-unit

test-unit:
	go test -v -count=1 ./...

test-race:
	go test -race -count=1 ./...

fmt:
	gofmt -s -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

vet:
	go vet ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

tidy-check:
	@cp go.mod go.mod.bak && cp go.sum go.sum.bak
	@go mod tidy
	@if ! diff -q go.mod go.mod.bak >/dev/null 2>&1 || ! diff -q go.sum go.sum.bak >/dev/null 2>&1; then \
		mv go.mod.bak go.mod; mv go.sum.bak go.sum; \
		echo "go.mod or go.sum is not tidy — run 'go mod tidy'"; \
		exit 1; \
	fi
	@mv go.mod.bak go.mod && mv go.sum.bak go.sum

check: fmt-check vet test-unit

release-check: check
	@if grep -q '^replace' go.mod; then \
		echo "ERROR: go.mod contains replace directives"; \
		grep '^replace' go.mod; \
		exit 1; \
	fi

release:
	@scripts/release.sh $(VERSION)

clean:
	rm -rf bin/
	go clean

tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
