# ctenter

A host-side tool that lists containers across different runtimes (Docker, containerd, CRI-O) and injects a shell agent to get interactive access — even in distroless or shell-less containers.   
 <img align="right" width="250" height="250" alt="ctenter_logo" src="https://github.com/user-attachments/assets/75b8cbe9-c1a8-4781-bad0-d037134e0127" />   



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

## How to do it manually (without ctenter) - limited

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
