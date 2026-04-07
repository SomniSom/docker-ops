# Docker Quick-ops — build and test (see readme.md)
.PHONY: all build install install-man clean test test-unit test-integration vet fmt lint gen-man goreleaser-check goreleaser-snapshot help

BINARY     ?= dq
OUT        ?= bin/$(BINARY)
MODULE     := github.com/SomniSom/docker-ops
VERSION    ?= dev
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
LDFLAGS    := -X '$(MODULE)/internal/version.Version=$(VERSION)'
ifneq ($(strip $(GIT_COMMIT)),)
LDFLAGS += -X '$(MODULE)/internal/version.Commit=$(GIT_COMMIT)'
endif

all: build

MANPREFIX ?= /usr/local/share/man

help:
	@echo "Targets: build install install-man gen-man goreleaser-check goreleaser-snapshot clean test test-unit test-integration vet fmt"

# Requires goreleaser v2: https://goreleaser.com/install/
goreleaser-check:
	goreleaser check

goreleaser-snapshot:
	goreleaser release --clean --snapshot

build:
	@mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT) ./cmd/dq

install:
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/dq

# Regenerate man/man1/*.1 (English); requires for packaging / install-man
gen-man:
	go run ./tools/genman

install-man: gen-man
	install -d $(DESTDIR)$(MANPREFIX)/man1
	install -m 644 man/man1/*.1 $(DESTDIR)$(MANPREFIX)/man1/

clean:
	rm -rf bin/
	rm -f man/man1/*.1

test: test-unit

# Default tests (no Docker required)
test-unit:
	go test ./... -short -count=1 -race

# Requires Docker; exercises docker compose
test-integration:
	go test ./... -tags=integration -count=1 -timeout=5m

vet:
	go vet ./...

fmt:
	go fmt ./...

lint: vet
	@echo "Add golangci-lint locally if desired: golangci-lint run"
