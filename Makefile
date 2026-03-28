.PHONY: build test test-v test-race test-integration test-live dev clean release fmt vet lint version

BINARY     = droidtether
BUILD_DIR  = ./build
CMD        = ./cmd/droidtether
GOARCH     = arm64
GOOS       = darwin
CGO        = 1

# ──────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────

build:
	@echo "→ Building $(BINARY) for $(GOOS)/$(GOARCH)..."
	GOARCH=$(GOARCH) GOOS=$(GOOS) CGO_ENABLED=$(CGO) \
		go build -ldflags="-X main.version=$(shell cat VERSIONS.md | grep '^## v' | head -1 | awk '{print $$2}')" \
		-o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY)"

build-intel:
	@echo "→ Building $(BINARY) for darwin/amd64..."
	GOARCH=amd64 GOOS=darwin CGO_ENABLED=1 \
		go build -o $(BUILD_DIR)/$(BINARY)-amd64 $(CMD)

clean:
	rm -rf $(BUILD_DIR)/
	@echo "✓ Cleaned"

# ──────────────────────────────────────────────
# Test
# ──────────────────────────────────────────────

test:
	go test ./...

test-v:
	go test -v ./...

test-race:
	go test -race ./...

test-integration:
	@echo "→ Integration tests (requires root for utun creation)"
	sudo go test -tags integration -v ./test/integration/...

test-live:
	@echo "→ Live end-to-end test (requires phone attached + USB tethering ON)"
	@echo "→ Running as sudo..."
	sudo bash scripts/test-live.sh

# ──────────────────────────────────────────────
# Development
# ──────────────────────────────────────────────

dev:
	@echo "→ Starting dev hot-reload loop (requires fswatch: brew install fswatch)"
	@bash scripts/dev-reload.sh

install-local: build
	@echo "→ Installing $(BINARY) to /usr/local/bin/ (requires sudo)"
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	sudo bash scripts/install-launchd.sh
	@echo "✓ Installed and daemon started"

uninstall-local:
	sudo bash scripts/uninstall-launchd.sh
	sudo rm -f /usr/local/bin/$(BINARY)
	@echo "✓ Uninstalled"

logs:
	tail -f /var/log/droidtether.log

daemon-status:
	launchctl list | grep droidtether || echo "droidtether daemon not running"

daemon-restart:
	sudo launchctl kickstart -k system/com.princePal.droidtether

# ──────────────────────────────────────────────
# Code Quality
# ──────────────────────────────────────────────

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null || (echo "Install: brew install golangci-lint" && exit 1)
	golangci-lint run

# ──────────────────────────────────────────────
# Release
# ──────────────────────────────────────────────

release: test test-race build
	@VERSION=$$(cat VERSIONS.md | grep '^## v' | head -1 | awk '{print $$2}'); \
	echo "→ Tagging release $$VERSION"; \
	git tag -a $$VERSION -m "Release $$VERSION"; \
	echo "→ Push tag with: git push origin $$VERSION"; \
	echo "→ GitHub Actions will build + publish the release"

version:
	@cat VERSIONS.md | grep '^## v' | head -1 | awk '{print $$2}'

# ──────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────

help:
	@echo "DroidTether — Make Targets"
	@echo ""
	@echo "  make build            Build arm64 binary"
	@echo "  make test             Run all unit tests"
	@echo "  make test-v           Run tests (verbose)"
	@echo "  make test-race        Run tests with race detector"
	@echo "  make test-integration Run integration tests (needs root)"
	@echo "  make test-live        Run live test (needs phone + root)"
	@echo "  make dev              Hot-reload dev loop"
	@echo "  make install-local    Install daemon locally for testing"
	@echo "  make uninstall-local  Remove local install"
	@echo "  make logs             Tail daemon log"
	@echo "  make daemon-restart   Restart the launchd daemon"
	@echo "  make fmt              Format all Go code"
	@echo "  make lint             Run golangci-lint"
	@echo "  make release          Tag + prepare release"
	@echo "  make version          Show current version"