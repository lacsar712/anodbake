# anodbake

Carbon anode baking furnace zone control. Operator HMI under `internal/web`.

## Requirements

- Go 1.22+ (container image uses golang:1.22)
- `GOTOOLCHAIN=local` recommended on host when using a pinned toolchain

## Build

```bash
export GOTOOLCHAIN=local
go build ./...
```

## Run

```bash
export GOTOOLCHAIN=local
go run ./cmd/steamboil
```

Open http://127.0.0.1:8080/

Default ingest secret: `dev-ingest-secret`.

## Test

```bash
export GOTOOLCHAIN=local
go test ./... -count=1
```

## Docker (benzhi)

Must build **linux/amd64** and **linux/arm64**:

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh anodbake linux/amd64
./build_benzhi_docker.sh anodbake linux/arm64
docker run -it anodbake:latest
# inside container:
export GOTOOLCHAIN=local
go version
go build ./...
go test ./... -count=1
```
