GO      := $(shell which go 2>/dev/null || echo /usr/local/go/bin/go)
BIN_DIR := bin
VERSION := v0.1.0

GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS_CTENTER  := -ldflags "-X github.com/g3rzi/ctenter/cmd.version=$(VERSION) -X github.com/g3rzi/ctenter/cmd.buildTime=$(BUILD_TIME) -X github.com/g3rzi/ctenter/cmd.gitCommit=$(GIT_COMMIT)"
LDFLAGS_CTENTERD := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

ifdef STATIC
  OUT_DIR          := $(BIN_DIR)/static
  CGO              := CGO_ENABLED=0
  STATIC_LDFLAGS   := -ldflags "-s -w -extldflags '-static' -X github.com/g3rzi/ctenter/cmd.version=$(VERSION) -X github.com/g3rzi/ctenter/cmd.buildTime=$(BUILD_TIME) -X github.com/g3rzi/ctenter/cmd.gitCommit=$(GIT_COMMIT)"
  STATIC_LDFLAGS_D := -ldflags "-s -w -extldflags '-static' -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"
else
  OUT_DIR          := $(BIN_DIR)
  CGO              :=
  STATIC_LDFLAGS   := $(LDFLAGS_CTENTER)
  STATIC_LDFLAGS_D := $(LDFLAGS_CTENTERD)
endif

.PHONY: all ctenter ctenterd submodule clean

all: ctenter ctenterd

$(OUT_DIR):
	mkdir -p $(OUT_DIR)

submodule:
	git submodule update --init --recursive

ctenter: $(OUT_DIR)
	$(CGO) $(GO) build $(STATIC_LDFLAGS) -o $(OUT_DIR)/ctenter .

ctenterd: submodule $(OUT_DIR)
	cd agent/ctenterd && $(CGO) $(GO) build $(STATIC_LDFLAGS_D) -o ../../$(OUT_DIR)/ctenterd .

clean:
	rm -rf $(BIN_DIR)
