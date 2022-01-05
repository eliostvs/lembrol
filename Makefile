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

## help: print this help message
.PHONY: help
help:
	echo 'Usage:'
	sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /' | sort

## clean: delete binary and development environment
.PHONY: clean
clean:
	rm $(DEV_MARKER) 2> /dev/null || true
	rm $(BINARY) 2> /dev/null || true
	rm coverage.out 2> /dev/null || true
	rm coverage.html 2> /dev/null || true

$(DEV_MARKER):
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(LINTER)
	touch $(DEV_MARKER)

## dev: prepare development environment
.PHONY: dev
dev: $(DEV_MARKER)

## deps/outdated: list outdated dependencies
.PHONY: deps/outdated
deps/outdated:
	go list -f "{{if and (not .Main) (not .Indirect)}} {{if .Update}} {{.Update}} {{end}} {{end}}" -m -u all 2> /dev/null | awk NF

## deps/upgrade: upgrade dependencies
.PHONY: deps/upgrade
deps/upgrade:
	go get -u $(pkg)
	go mod tidy
	go mod verify

## build: create binary
.PHONY: build
build: dev
	echo "version: $(VERSION)"
	CGO_ENABLED=0 GOARCH=amd64 go build -o $(BINARY) -ldflags '$(LDFLAGS)' $(BINARY_DIR)

## run [args]: run app in development mode
.PHONY: run
run: dev
	go run $(BINARY_DIR) $(args)

## format: format files
.PHONY: format
format: dev
	goimports -l -w .

## test/lint: run lint
.PHONY: test/lint
test/lint: dev
	golangci-lint run

## test [args] [pkg]: run tests
.PHONY: test
test: dev
	go test $(args) -v -race -cover -coverprofile=coverage.out $(pkg)

## test/all: run lint and tests
.PHONY: test/all
test/all: test/lint test

coverage.out: test

## test/report: shows coverage report
.PHONY: test/report
test/report: coverage.out
	go tool cover -html=coverage.out -o coverage.html
ifeq ($(OSFLAG),Linux)
	xdg-open coverage.html
endif
ifeq ($(OSFLAG),Darwin)
	open coverage.html
endif
