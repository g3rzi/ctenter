GO      := $(shell which go 2>/dev/null || echo /usr/local/go/bin/go)
BIN_DIR := bin
VERSION := v0.1.0

# Target platform (override with OS=linux ARCH=arm64 make release)
OS   ?= linux
ARCH ?= amd64

GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS_CTENTER  := -ldflags "-X github.com/g3rzi/ctenter/cmd.version=$(VERSION) -X github.com/g3rzi/ctenter/cmd.buildTime=$(BUILD_TIME) -X github.com/g3rzi/ctenter/cmd.gitCommit=$(GIT_COMMIT)"
LDFLAGS_CTENTERD := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

LDFLAGS_CTENTER_STATIC  := -ldflags "-s -w -extldflags '-static' -X github.com/g3rzi/ctenter/cmd.version=$(VERSION) -X github.com/g3rzi/ctenter/cmd.buildTime=$(BUILD_TIME) -X github.com/g3rzi/ctenter/cmd.gitCommit=$(GIT_COMMIT)"
LDFLAGS_CTENTERD_STATIC := -ldflags "-s -w -extldflags '-static' -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

ifdef STATIC
  OUT_DIR := $(BIN_DIR)/static
  CGO     := CGO_ENABLED=0
  ACTIVE_LDFLAGS_CTENTER  := $(LDFLAGS_CTENTER_STATIC)
  ACTIVE_LDFLAGS_CTENTERD := $(LDFLAGS_CTENTERD_STATIC)
else
  OUT_DIR := $(BIN_DIR)
  CGO     :=
  ACTIVE_LDFLAGS_CTENTER  := $(LDFLAGS_CTENTER)
  ACTIVE_LDFLAGS_CTENTERD := $(LDFLAGS_CTENTERD)
endif

# Release output directory
RELEASE_DIR := dist

.PHONY: all ctenter ctenterd submodule clean release release-ctenter release-ctenter-static release-ctenterd release-ctenterd-static

all: ctenter ctenterd

$(OUT_DIR):
	mkdir -p $(OUT_DIR)

submodule:
	git submodule update --init --recursive

ctenter: $(OUT_DIR)
	$(CGO) GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(ACTIVE_LDFLAGS_CTENTER) -o $(OUT_DIR)/ctenter .

ctenterd: submodule $(OUT_DIR)
	cd agent/ctenterd && $(CGO) GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(ACTIVE_LDFLAGS_CTENTERD) -o ../../$(OUT_DIR)/ctenterd .

# ── Release targets ──────────────────────────────────────────────────────────

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)

# ctenter dynamic
release-ctenter: $(RELEASE_DIR)
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(LDFLAGS_CTENTER) -o $(RELEASE_DIR)/ctenter-$(OS)-$(ARCH) .
	tar -czf $(RELEASE_DIR)/ctenter-$(OS)-$(ARCH).tar.gz -C $(RELEASE_DIR) ctenter-$(OS)-$(ARCH)

# ctenter static
release-ctenter-static: $(RELEASE_DIR)
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(LDFLAGS_CTENTER_STATIC) -o $(RELEASE_DIR)/ctenter-$(OS)-$(ARCH)-static .
	tar -czf $(RELEASE_DIR)/ctenter-$(OS)-$(ARCH)-static.tar.gz -C $(RELEASE_DIR) ctenter-$(OS)-$(ARCH)-static

# ctenterd dynamic
release-ctenterd: submodule $(RELEASE_DIR)
	cd agent/ctenterd && GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(LDFLAGS_CTENTERD) -o ../../$(RELEASE_DIR)/ctenterd-$(OS)-$(ARCH) .
	tar -czf $(RELEASE_DIR)/ctenterd-$(OS)-$(ARCH).tar.gz -C $(RELEASE_DIR) ctenterd-$(OS)-$(ARCH)

# ctenterd static
release-ctenterd-static: submodule $(RELEASE_DIR)
	cd agent/ctenterd && CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(LDFLAGS_CTENTERD_STATIC) -o ../../$(RELEASE_DIR)/ctenterd-$(OS)-$(ARCH)-static .
	tar -czf $(RELEASE_DIR)/ctenterd-$(OS)-$(ARCH)-static.tar.gz -C $(RELEASE_DIR) ctenterd-$(OS)-$(ARCH)-static

# Build all release artifacts
release: release-ctenter release-ctenter-static release-ctenterd release-ctenterd-static
	@echo ""
	@echo "Release artifacts in $(RELEASE_DIR)/:"
	@ls -lh $(RELEASE_DIR)/*.tar.gz

clean:
	rm -rf $(BIN_DIR) $(RELEASE_DIR)
