.PHONY: build test vet fmt fmt-check tidy tidy-check check clean release-check release-snapshot

BINARY ?= classreach
PKG := ./...
GORELEASER_VERSION ?= $(shell cat .goreleaser-version)
GORELEASER := CGO_ENABLED=0 go run github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

build:
	mkdir -p bin
	go build -trimpath -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test -count=1 $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

tidy:
	go mod tidy

fmt-check:
	@test -z "$$(gofmt -l cmd internal)"

tidy-check:
	go mod tidy -diff

check: fmt-check tidy-check vet test build

release-check:
	$(GORELEASER) check

release-snapshot:
	$(GORELEASER) release --snapshot --clean

clean:
	rm -rf bin dist
