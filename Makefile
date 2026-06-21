.PHONY: build install test race smoke vet fmt

# Build the gw binary into ./bin (CGO-free, matches scripts/smoke.sh).
build:
	CGO_ENABLED=0 go build -o bin/gw ./cmd/gw

# Install gw onto $GOBIN/$GOPATH/bin so `gw` is on PATH for humans and agents.
install:
	CGO_ENABLED=0 go install ./cmd/gw

test:
	CGO_ENABLED=0 go test ./...

# Concurrency paths need the race detector, which requires cgo.
race:
	CGO_ENABLED=1 go test -race ./...

smoke:
	bash scripts/smoke.sh

vet:
	CGO_ENABLED=0 go vet ./...

fmt:
	gofmt -l .
