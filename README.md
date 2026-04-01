# Kanshi Agent

This repository contains the **Kanshi Agent**, a small system monitoring agent written in Go.

The agent's primary role is to collect system metrics and ship them to the [Kanshi Core](https://github.com/kanshi-dev/core) service, which handles storage, indexing, and visualization.

---

## What It Does

- Collects system stats using `gopsutil`
- Batches metrics in memory
- Flushes by size or time interval
- Sends data to a gRPC endpoint
- Stateless, best-effort delivery

Under the hood, it’s essentially a thin wrapper around the excellent [`gopsutil`](https://github.com/shirou/gopsutil) library with a simple batching + transport layer.

---

## Architecture (Simple by Design)

collect → batch → send

---

## Run

```bash
go run cmd/agent/main.go
```

## Configure via environment variables:
| Variable | Default | Description |
|---|---|---|
| `KANSHI_CORE_ADDR` | `127.0.0.1:50051` | The gRPC address of the Kanshi core service. |
| `KANSHI_API_KEY` | (empty) | API key for authentication with the core service. |
| `KANSHI_INTERVAL` | `5s` | How often the agent collects system metrics. |
| `KANSHI_BATCH_MAX` | `100` | Maximum number of points to batch before flushing. |
| `KANSHI_FLUSH_EVERY` | `10s` | Maximum time to wait before flushing regardless of batch size. |
| `KANSHI_HOST_TAGS` | (empty) | Comma-separated tags for the host (e.g. `env:prod,region:us-west`). |

```bash
# Example
export KANSHI_CORE_ADDR=localhost:50051
export KANSHI_INTERVAL=5s
go run cmd/agent/main.go
```


### Why This Exists?

- To understand how monitoring agents work internally
- To practice Go project structure
- To build toward a larger system ([Kanshi Core](https://github.com/kanshi-dev/core))

Future versions may add retries, streaming, and reliability features — but v1 stays intentionally simple.