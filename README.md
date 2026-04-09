<p align="center">
  <img src="https://github.com/user-attachments/assets/75b8cbe9-c1a8-4781-bad0-d037134e0127" width="220"/>
</p>
<p align="center">
  A host-side tool to access <b>any container</b> — even without a shell.
</p>
<p align="center">
  <img src="https://img.shields.io/github/v/release/g3rzi/ctenter?style=for-the-badge" />
  <img src="https://img.shields.io/github/downloads/g3rzi/ctenter/total?style=for-the-badge" />
  <img src="https://img.shields.io/github/license/g3rzi/ctenter?style=for-the-badge" />
  <img src="https://img.shields.io/badge/go-1.21%2B-blue?style=for-the-badge&logo=go" />
</p>

A host-side tool that lists containers across runtimes (Docker, containerd, CRI-O) and gives you interactive access — even in distroless or shell-less containers.  

## ✨ Features 
🔍 Cross-runtime discovery — Docker, containerd, CRI-O  
🐚 Shell access without a shell — works in distroless containers  
⚡ One-shot command execution — `--exec`  
🧩 Custom agent support — bring your own binary  
🪶 Lightweight injection via `/proc/<pid>/root`  
🔐 No container modification required  
🧠 Namespace-aware execution using `nsenter`  

## ⚡ Quick start
```bash
# List containers
sudo ctenter list

# Enter container
sudo ctenter --pid <PID>
```

## Why I built this

Modern containers are often built to be minimal (e.g. distroless images) and intentionally exclude shells like `/bin/sh` or `/bin/bash`.

While this improves security and reduces image size, it makes debugging and inspection much harder.

Trying to exec into such containers usually fails with errors like:
```
OCI runtime exec failed: exec failed: unable to start container process: exec: "sh": executable file not found in $PATH
```

This creates a frustrating situation:
- You can’t `docker exec -it <container> sh`
- You can’t easily inspect the filesystem or running processes
- Debugging production issues becomes significantly harder

`ctenter` was built to solve this problem by allowing interactive access to *any* container — even those without a shell — by injecting a lightweight agent directly into the container’s namespaces.  

## Demo  


https://github.com/user-attachments/assets/82af3db4-a2fb-41ec-b634-a1673f28b00f


## How it works

`ctenter` works in two steps:

1. Injects a small static agent binary (`ctenterd`) into the container's filesystem via `/proc/<pid>/root`
2. Uses `nsenter` to enter the container's namespaces and execute the agent

This means you can get a shell in any container regardless of whether it has `/bin/sh`, `/bin/bash`, or any other shell installed.


## How to do it manually (without ctenter)

It is possible to access a container’s filesystem and namespaces manually using standard Linux tools but it’s tedious and error-prone.

### 1. Find the container PID

For Docker:  
```bash
docker inspect -f '{{.State.Pid}}' <container>
```

For containerd / CRI-based runtimes, you can use:  
```
crictl inspect <container-id> | grep pid
```
### 2. Explore the container filesystem  

Once you have the PID, you can access the container’s root filesystem via `/proc`:  
```
ls /proc/<PID>/root
```
This gives you direct visibility into the container’s filesystem from the host.  

### 3. Enter the container namespaces  
You can use `nsenter` to enter the container’s namespaces:  
```
sudo nsenter -t <PID> -m -u -i -n -p
```

However, you’ll still hit the same limitation:  
```
nsenter: failed to execute /bin/bash: No such file or directory
```
If the container doesn’t include a shell, you won’t be able to get interactive access.  

### 4. Workaround: inject your own binary   
A manual workaround is to copy a statically linked binary (e.g. `busybox`) into the container filesystem via `/proc/<PID>/root`, and then execute it inside the container namespaces.

Install a static busybox on the host:
```
sudo apt install busybox-static 
```

Copy it into the container:  
```
sudo cp /bin/busybox /proc/<PID>/root/tmp/busybox
sudo chmod +x /proc/<PID>/root/tmp/busybox
```

Then execute it inside the container namespaces:  
```
sudo nsenter -t <PID> -m -u -i -n -p --root=/proc/<PID>/root /tmp/busybox sh
```
This can give you a shell even if the container doesn’t include one.   

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

# Cross-compile for a specific OS/architecture (defaults: linux/amd64)
make OS=linux ARCH=arm64

# Build individually
make ctenter
make ctenterd

# Clean build output
make clean
```

## Release

Produces versioned, platform-named tarballs in `dist/`:

```bash
# Build all release artifacts (linux/amd64 by default)
make release

# Cross-compile release artifacts for arm64
make release OS=linux ARCH=arm64
```

Output files:

| File | Description |
|------|-------------|
| `ctenter-linux-amd64.tar.gz` | Host tool (dynamic) |
| `ctenter-linux-amd64-static.tar.gz` | Host tool (fully static) | 
| `ctenterd-linux-amd64.tar.gz` | Agent binary (dynamic) |  
| `ctenterd-linux-amd64-static.tar.gz` | Agent binary (static, for injection) |

Individual targets:

```bash
make release-ctenter           # dynamic host tool only
make release-ctenter-static    # static host tool only
make release-ctenterd          # dynamic host tool only
make release-ctenterd-static   # static agent only
```

## Test no-shell containers

Use the example image in `examples/` to validate `ctenter` against a container that has no `/bin/sh`:

```bash
docker build -f examples/Dockerfile.noshell -t ctenter-noshell-test .
docker run -d --name ctenter-noshell-test ctenter-noshell-test
```

Full test instructions are in `examples/README.md`.

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

# Force a specific runtime (auto, docker, cri — default: auto)
sudo ctenter list --runtime docker
sudo ctenter list --runtime cri
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
│   └── ctenterd/          # Shell agent (git submodule → github.com/g3rzi/ctenterd)
│       ├── builtin/       # Built-in shell commands (ls, ps, cat, ping, ...)
│       ├── internal/      # Agent-internal utilities
│       └── shell/         # Shell REPL runner
├── cmd/
│   ├── list/              # `ctenter list` subcommand
│   └── root.go            # Root command and shell entry point
├── examples/
│   ├── Dockerfile.noshell # Scratch-based test image (no shell)
│   └── README.md          # Test instructions
├── pkg/
│   ├── discover/          # Container runtime discovery (Docker, CRI)
│   ├── inject/            # Agent injection via /proc/<pid>/root
│   ├── nsenter/           # Namespace entry
│   └── shell/             # Interactive and exec shell helpers
├── dist/                  # Release tarballs (generated by make release)
├── main.go                # ctenter entry point
├── Makefile
├── go.mod
└── go.sum
```

## License

MIT
