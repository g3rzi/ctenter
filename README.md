# ctenter

A host-side tool that lists containers across different runtimes (Docker, containerd, CRI-O) and injects a shell agent to get interactive access — even in distroless or shell-less containers.

## How it works

`ctenter` works in two steps:

1. Injects a small static agent binary (`ctenterd`) into the container's filesystem via `/proc/<pid>/root`
2. Uses `nsenter` to enter the container's namespaces and execute the agent

This means you can get a shell in any container regardless of whether it has `/bin/sh`, `/bin/bash`, or any other shell installed.

## Requirements

- Linux host
- Root privileges (`sudo`)
- Go 1.21+ (to build)

## Clone

This repo uses a Git submodule for the `ctenterd` agent. Clone with:

```bash
git clone --recurse-submodules https://github.com/g3rzi/ctenter.git
```

If you already cloned without the flag and `agent/ctenterd` is empty:

```bash
git submodule update --init --recursive
```

## Build

```bash
# Build both binaries into bin/
make

# Build static binaries into bin/static/
make STATIC=1

# Build individually
make ctenter
make ctenterd

# Clean build output
make clean
```

## Update submodule
### Pull latest ctenterd
```
cd agent\ctenterd
git pull origin main
```

### Go back to ctenter root and commit the updated pointer
```
cd ..\..
git add agent\ctenterd
git commit -m "update ctenterd submodule to latest"
git push -u origin main
```

## Usage

### List containers

```bash
sudo ctenter list

# Wide output (includes pod ID, container ID, image ID)
sudo ctenter list --wide

# Untruncated fields
sudo ctenter list --no-trunc
```

### Get an interactive shell

```bash
# Get the container PID from `ctenter list`, then:
sudo ctenter shell --pid <PID>

# Or using the root shorthand
sudo ctenter --pid <PID>
```

### Run a one-shot command

```bash
sudo ctenter shell --pid <PID> --exec "ls -la /etc"
```

### Use a custom agent binary

```bash
sudo ctenter shell --pid <PID> --agent-path /path/to/custom-agent
```

### Version

```bash
ctenter --version
ctenterd --version
```

## Project layout

```
ctenter/
├── agent/
│   └── ctenterd/          # Shell agent injected into containers
│       ├── builtin/       # Built-in shell commands (ls, ps, cat, ...)
│       ├── internal/      # Agent-internal utilities
│       └── shell/         # Shell runner
├── cmd/
│   ├── list/              # `ctenter list` subcommand
│   └── root.go            # Root command and shell entry point
├── pkg/
│   ├── color/             # Terminal color helpers
│   ├── discover/          # Container runtime discovery (CRI)
│   ├── inject/            # Agent injection via /proc/<pid>/root
│   ├── nsenter/           # Namespace entry
│   └── shell/             # Interactive and exec shell helpers
├── main.go                # ctenter entry point
├── Makefile
├── go.mod
└── go.sum
```

## License

MIT
