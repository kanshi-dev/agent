# Kanshi Agent

[![CI](https://github.com/kanshi-dev/agent/actions/workflows/ci.yaml/badge.svg)](https://github.com/kanshi-dev/agent/actions/workflows/ci.yaml)

Kanshi Agent is a small Go service that collects CPU, memory, and disk usage with gopsutil, batches the points in memory, and sends them to Kanshi Core over authenticated gRPC.

```text
collect -> batch -> send -> reconnect and retry when needed
```

## Install

The current stable release is `v1.0.0`:

```sh
curl -fsSL https://kanshi.dev/install.sh |
  KANSHI_VERSION=v1.0.0 sh
```

Run it with a core address and ingest key:

```sh
export KANSHI_CORE_ADDR=your-server:50051
export KANSHI_API_KEY=your-ingest-key
kanshi-agent
```

For systemd Linux:

```sh
curl -fsSL https://kanshi.dev/install.sh |
  sudo KANSHI_VERSION=v1.0.0 \
  KANSHI_CORE_ADDR=your-server:50051 \
  KANSHI_API_KEY=your-ingest-key \
  sh -s -- --systemd
```

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `KANSHI_CORE_ADDR` | `127.0.0.1:50051` | Core gRPC address |
| `KANSHI_API_KEY` | empty | Shared ingest key required by core |
| `KANSHI_INTERVAL` | `5s` | Collection cadence |
| `KANSHI_BATCH_MAX` | `100` | Size-triggered flush threshold |
| `KANSHI_FLUSH_EVERY` | `10s` | Time-triggered flush |
| `KANSHI_HOST_TAGS` | empty | Comma-separated host tags |
| `KANSHI_LOG_LEVEL` | `info` | Log level |

The agent ID persists in `.kanshi-id` in the working directory, or in `/var/lib/kanshi-agent` for the packaged systemd service.

## Develop

```sh
go run ./cmd/agent
go test ./...
go vet ./...
go build ./...
```

See the [canonical quickstart](https://github.com/kanshi-dev/core/blob/main/QUICKSTART.md) for the complete stack.

## Support and security

Use GitHub issues for public support. Report vulnerabilities through [private vulnerability reporting](SECURITY.md). Kanshi follows semantic versioning from `v1.0.0`.
