# container examples

Demonstrates the TAU container library — OCI-aligned runtime abstraction (`tau/container`) and its runtime sub-modules. Each example exercises the published contract a downstream consumer would actually import.

## Prerequisites

- Docker daemon running and reachable (`docker info` succeeds).
- Pull the base image any demo needs ahead of time. The current set uses `alpine:3.21`:
  ```bash
  docker pull alpine:3.21
  ```

## Examples

| Example | Demonstrates | Run Command |
|---------|--------------|-------------|
| docker-hello | Lifecycle (Create/Start/Stop/Remove), one-shot Exec, Inspect with the nil-manifest fallback path | `go run ./container/docker-hello/` |

---

### docker-hello

End-to-end smoke demo of the `tau/container` runtime contract against the Docker Engine API. Walks through the full lifecycle, captures stdout from a one-shot `Exec`, and demonstrates the `container.Fallback()` substitution that callers do when an image carries no `/etc/tau/manifest.json`.

**Run:**
```bash
go run ./container/docker-hello/
```

See [docker-hello/README.md](./docker-hello/README.md) for the per-demo walkthrough, expected output, and the inspection commands that surface the reserved `tau.*` labels.

## References

- Library docs and source: [`tailored-agentic-units/container`](https://github.com/tailored-agentic-units/container)
- Docker runtime sub-module: [`tailored-agentic-units/container/docker`](https://github.com/tailored-agentic-units/container/tree/main/docker)
