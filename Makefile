ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
endif

.DEFAULT_GOAL  = help
.DELETE_ON_ERROR:
.ONESHELL:
.SHELLFLAGS    := -eu -o pipefail -c
.SILENT:
MAKEFLAGS      += --no-builtin-rules
MAKEFLAGS      += --warn-undefined-variables
SHELL          = bash

BINARY         = remember
BINARY_DIR     = ./cmd/$(BINARY)
DEV_MARKER     = .__dev
LDFLAGS        += -X "main.Version=${VERSION}"
LINTER         = v1.42.1
OSFLAG         ?=
VERSION        ?= $(shell git describe --tags $(shell git rev-list --tags --max-count=1) 2>/dev/null || echo "dev")
args           ?=
pkg            ?=./...

ifeq ($(OS),Windows_NT)
	OSFLAG = "windows"
else
	OSFLAG = $(shell uname -s)
endif

.PHONY: help
help:
	$(info Available tasks:)
	$(info | build              Create binary)
	$(info | clean              Delete binary and development files)
	$(info | dev                Download development dependencies)
	$(info | format             Format files using goimports)
	$(info | help               Show this help message)
	$(info | lint               Run lint)
	$(info | outdated           List outdated dependencies)
	$(info | run [args]         Run app in development mode)
	$(info | test [args] [pkg]  Run tests)
	$(info | test-all           Run lint and tests)
	$(info | test-report        Open coverage report)
	$(info | upgrade [pkg]      Upgrade dependencies)

.PHONY: clean
clean:
	rm $(DEV_MARKER)
	rm $(BINARY) 2> /dev/null || true
	rm coverage.out 2> /dev/null || true
	rm coverage.html 2> /dev/null || true

$(DEV_MARKER):
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(LINTER)
	touch $(DEV_MARKER)

.PHONY: dev
dev: $(DEV_MARKER)

.PHONY: outdated
outdated:
	go list -f "{{if and (not .Main) (not .Indirect)}} {{if .Update}} {{.Update}} {{end}} {{end}}" -m -u all 2> /dev/null | awk NF

.PHONY: upgrade
upgrade:
	go get -u $(pkg)
	go mod tidy

.PHONY: build
build: dev
	echo "version: $(VERSION)"
	CGO_ENABLED=0 GOARCH=amd64 go build -o $(BINARY) -ldflags '$(LDFLAGS)' $(BINARY_DIR)

.PHONY: run
run: dev
	go run $(BINARY_DIR) $(args)

.PHONY: format
format: dev
	goimports -l -w .

.PHONY: lint
lint: dev
	golangci-lint run

.PHONY: test
test: dev
	go test $(args) -v -race -cover -coverprofile=coverage.out $(pkg)

.PHONY: test-all
test-all: lint test

.PHONY: test-report
test-report: test
	go tool cover -html=coverage.out -o coverage.html
ifeq ($(OSFLAG),Linux)
	xdg-open coverage.html
endif
ifeq ($(OSFLAG),Darwin)
	open coverage.html
endif
