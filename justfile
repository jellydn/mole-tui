set shell := ["/bin/sh", "-eu", "-c"]

# Build version. Override with `just build VERSION=v0.1.0`.
VERSION := "dev"

# Build ./bin/mole-tui for the host OS/arch.
build:
    mkdir -p bin
    go build -ldflags '-X main.version={{VERSION}}' -trimpath -o bin/mole-tui ./cmd/mole-tui

# go install ./cmd/mole-tui to $GOBIN.
install:
    go install -ldflags '-X main.version={{VERSION}}' -trimpath ./cmd/mole-tui

# Run all unit tests.
test:
    go test ./...

# Run go vet on all packages.
vet:
    go vet ./...

# Check gofmt (fails if any file is unformatted).
fmt:
    @test -z "$(gofmt -l -s .)"

# Build and run.
dev: build
    ./bin/mole-tui

# CI: fmt → vet → test → build.
ci: fmt vet test build
