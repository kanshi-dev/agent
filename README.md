# Kanshi Agent

Kanshi Agent is a small system monitoring agent written in Go.

It collects basic system metrics (CPU, memory, disk, network), batches them, and sends them to a core service over gRPC.

This is a learning / pet project focused on building a clean monitoring pipeline from scratch.

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

No retries, no streaming, no disk buffering — just a minimal working agent.

---

## Run

```bash
go run cmd/agent/main.go
```

## Configure via environment variables:

```yaml
CORE_ADDR=localhost:50051
INTERVAL=5s
BATCH_MAX=100
FLUSH_EVERY=10s

```


### Why This Exists?

	-	To understand how monitoring agents work internally
	-	To practice Go project structure
	-	To build toward a larger system (Kanshi Core)

Future versions may add retries, streaming, and reliability features — but v1 stays intentionally simple.