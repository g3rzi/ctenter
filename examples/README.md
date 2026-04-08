# No-Shell Container Test Image

This folder contains a minimal Docker image used to test `ctenter` against a container with no shell and no common userland tools.

The image in `Dockerfile.noshell` uses `scratch` as the runtime base, so commands like `/bin/sh` or `bash` do not exist in the container.

## Why this exists

`ctenter` is designed to provide shell access even for distroless or shell-less containers by injecting and running `ctenterd`.

This image is a simple validation target for that workflow.

## Build the image

From the repository root:

```bash
docker build -f examples/Dockerfile.noshell -t ctenter-noshell-test .
```

## Run the container

```bash
docker run -d --name ctenter-noshell-test ctenter-noshell-test
```

## Confirm there is no shell

This should fail, which is expected:

```bash
docker exec -it ctenter-noshell-test /bin/sh
```

## Test with ctenter

```bash
# Option 1: use ctenter list and pick PID
sudo ./bin/ctenter list

# Option 2: get PID directly from Docker
docker inspect ctenter-noshell-test --format '{{.State.Pid}}'

# Then enter with ctenter
sudo ./bin/ctenter shell --pid <PID>
```

## Cleanup

```bash
docker rm -f ctenter-noshell-test
```
