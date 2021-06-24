.DELETE_ON_ERROR:
.ONESHELL:
.SILENT:
.SHELLFLAGS    := -eu -o pipefail -c
.DEFAULT_GOAL  = help
MAKEFLAGS      += --no-builtin-rules
MAKEFLAGS      += --warn-undefined-variables
SHELL          = bash

ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
endif

BINARY         = remember
LDFLAGS        += -X "main.version=${VERSION}"
LINTER_VERSION = v1.38.0
VERSION        ?= $(shell git describe --tags $(shell git rev-list --tags --max-count=1) 2>/dev/null || echo "dev")
args           ?=
pkg            ?=./...

.PHONY: help
help:
	$(info Available targets:)
	$(info | build              Create binary)
	$(info | clean              Delete binary and temp files)
	$(info | format             Format files using goimports)
	$(info | help               Show this help message)
	$(info | outdated           List outdated dependencies)
	$(info | install            Download dependencies)
	$(info | install-all        Download dependencies and lint)
	$(info | lint               Run lint)
	$(info | run [args]         Run app in development mode)
	$(info | test [args] [pkg]  Run tests)
	$(info | test-all           Run lint and tests)
	$(info | test-report        Open coverage report)
	$(info | tidy               Add missing and remove unused dependencies)
	$(info | tools              Download tools)
	$(info | upgrade [pkg]      Upgrade dependencies)

.PHONY: clean
clean:
	rm $(BINARY) 2> /dev/null || true
	rm coverage.out 2> /dev/null || true
	rm coverage.html 2> /dev/null || true

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: install
install:
	go mod download

.PHONY: tools
tools:
	go get golang.org/x/tools/cmd/goimports
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(LINTER_VERSION)

.PHONY: outdated
outdated:
	go list -f "{{if and (not .Main) (not .Indirect)}} {{if .Update}} {{.Update}} {{end}} {{end}}" -m -u all 2> /dev/null | awk NF

.PHONY: upgrade
upgrade:
	go get -u $(pkg)
	go mod tidy

.PHONY: build
build:
	echo "version: $(VERSION)"
	CGO_ENABLED=0 GOARCH=amd64 go build -o $(BINARY) -ldflags '$(LDFLAGS)' ./cmd/remember

.PHONY: run
run: 
	go run ./cmd/remember -log debug.log $(args)

.PHONY: format
format:
	goimports -l -w .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test $(args) -v -race -cover -coverprofile=coverage.out $(pkg)
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-report
test-report: clean test
	xdg-open coverage.html

.PHONY: test-all
test-all: lint test
