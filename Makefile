.PHONY: build install test race smoke vet fmt sizecheck web-build

# Guardrail: the web UI will be embedded into the binary via go:embed (ADR 0042),
# so warn when gw inflates past this. Override to tune: make build MAX_BIN_MB=120.
MAX_BIN_MB ?= 100

# Build the gw binary into ./bin (CGO-free, matches scripts/smoke.sh).
build:
	CGO_ENABLED=0 go build -o bin/gw ./cmd/gw
	@$(MAKE) --no-print-directory sizecheck BIN=bin/gw

# sizecheck warns (does not fail) when BIN exceeds the MAX_BIN_MB guardrail.
sizecheck:
	@bytes=$$(wc -c < "$(BIN)" | tr -d ' '); mb=$$(( bytes / 1048576 )); \
	echo "gw binary: $${mb} MB ($${bytes} bytes)"; \
	if [ "$${mb}" -gt "$(MAX_BIN_MB)" ]; then \
		printf 'WARNING: gw binary is %s MB, over the %s MB guardrail — check embedded web assets / dependencies (ADR 0042).\n' "$${mb}" "$(MAX_BIN_MB)" >&2; \
	fi

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

# Rebuild the static SPA that is committed and embedded into the gw binary.
web-build:
	npm --prefix web ci
	npm --prefix web run build
