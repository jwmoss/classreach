.PHONY: build test vet fmt fmt-check tidy tidy-check check clean release-tool-check release-check release-snapshot test-e2e

BINARY ?= classreach
PKG := ./...
GORELEASER_VERSION := $(shell cat .goreleaser-version)
GORELEASER := goreleaser

build:
	mkdir -p bin
	go build -trimpath -o bin/$(BINARY) ./cmd/$(BINARY)

test test-e2e: build
	go test -tags=live -count=1 -timeout=2m ./tests

vet:
	go vet -tags=live $(PKG)

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

tidy:
	go mod tidy

fmt-check:
	@test -z "$$(gofmt -l cmd internal tests)"

tidy-check:
	go mod tidy -diff

check: fmt-check tidy-check vet build

release-tool-check:
	@$(GORELEASER) --version | grep -Eq "^GitVersion: +v?$(subst .,[.],$(GORELEASER_VERSION:v%=%))$$" || { echo "Install GoReleaser $(GORELEASER_VERSION)"; exit 1; }

release-check: release-tool-check
	$(GORELEASER) check

release-snapshot: release-tool-check
	$(GORELEASER) release --snapshot --clean

clean:
	trash bin dist
