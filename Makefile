VERSION    := $(shell git describe --tags --match "v[0-9]*" --abbrev=7 --always --dirty 2>/dev/null | sed 's/^$$/v0.0.0-dev/')
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILT_AT   := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "\
  -X main.Version=$(VERSION) \
  -X main.Commit=$(COMMIT) \
  -X main.BuiltAt=$(BUILT_AT)"

BINARY  := bin/fusemomo
PACKAGE := github.com/fusemomo/fusemomo-cli

.PHONY: build test test-unit lint install clean release

## build: Compile the fusemomo binary with version/commit/date injection
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY) .
	@echo "Built $(BINARY) $(VERSION) ($(COMMIT))"

## test: Run all tests with race detector
test:
	go test -race ./...

## test-unit: Run unit tests only
test-unit:
	go test -race ./test/... -v

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## install: Install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) .

## clean: Remove the bin/ directory
clean:
	rm -rf bin/

## release: Build cross-platform releases with GoReleaser
release:
	goreleaser release --clean

## run: Run directly with go run
run:
	go run $(LDFLAGS) . $(ARGS)

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
