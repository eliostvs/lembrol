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

BINARY         = lembrol
BINARY_DIR     = ./cmd/$(BINARY)
DEV_MARKER     = .__dev
LINTER         = v1.55.2
OSFLAG         ?=
PROFILE_COV    = coverage.cov
REPORT_COV     = coverage.html
VERSION        = `cat .bumpversion.cfg | grep current_version | awk '{print $$3}'`
args           ?=
pkg            ?=./...

ifeq ($(OS),Windows_NT)
	OSFLAG = "windows"
else
	OSFLAG = $(shell uname -s)
endif

## help: print this help message
help:
	echo 'Usage:'
	sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /' | sort
.PHONY: help

## clean: delete binary and development environment
clean:
	rm $(DEV_MARKER) 2> /dev/null || true
	rm coverage* 2> /dev/null || true
	rm *.log 2> /dev/null || true
.PHONY: clean

$(DEV_MARKER):
	go mod download
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(LINTER)
	touch $(DEV_MARKER)

## dev: prepare development environment
dev: $(DEV_MARKER)
.PHONY: dev

## deps-outdated: list outdated dependencies
deps-outdated:
	go list -f "{{if and (not .Main) (not .Indirect)}} {{if .Update}} {{.Update}} {{end}} {{end}}" -m -u all 2> /dev/null | awk NF
.PHONY: deps-outdated

## deps-tidy: remove unused and check hash of the dependencies
deps-tidy:
	go mod tidy
	go mod verify
.PHONY: deps-tidy

## deps-upgrade [pkg]: upgrade dependencies
deps-upgrade: deps-tidy
	go get -u $(pkg)
.PHONY: deps-upgrade

## build: create snapshot release
build: dev
	go run github.com/goreleaser/goreleaser@latest build --clean --snapshot --single-target
.PHONY: build

## run [args]: run app in development mode
run: dev
	go run $(BINARY_DIR) --log $(args)
.PHONY: run

## format: format files
format: dev
	go run golang.org/x/tools/cmd/goimports@latest -l -w .
	go run mvdan.cc/gofumpt@latest -l -w .
.PHONY: format

## lint: run lint
lint: dev
	golangci-lint run
.PHONY: lint

## test [args] [pkg]: run unit tests
test: dev
	go test $(args) -race -shuffle=on -cover -coverprofile=${PROFILE_COV} $(pkg)
.PHONY: test

## vulncheck [pkg]: run vulnerability check
vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ${pkg}
.PHONY: vulncheck

## test-all: run lint and tests
test-all: lint test vulncheck
.PHONY: test-all

${PROFILE_COV}: test

## test-report: shows coverage report
test-report: ${PROFILE_COV}
	go tool cover -html=${PROFILE_COV} -o ${REPORT_COV}
ifeq ($(OSFLAG),Linux)
	xdg-open ${REPORT_COV}
endif
ifeq ($(OSFLAG),Darwin)
	open ${REPORT_COV}
endif
.PHONY: test-report

release-%:
	git flow init -d
	@grep -q '\[Unreleased\]' CHANGELOG.md || (echo 'Create the [Unreleased] section in the changelog first!' && exit)
	bumpversion --verbose --commit $*
	git flow release start $(VERSION)
	GIT_MERGE_AUTOEDIT=no git flow release finish -m "Merge branch release/$(VERSION)" $(VERSION) -p
