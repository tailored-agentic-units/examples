# docker-hello

End-to-end smoke demo of the `tau/container` runtime contract against the Docker Engine API. Walks through the full lifecycle (Create → Start → Exec → Inspect → Stop → Remove), then explicitly demonstrates the nil-manifest fallback substitution that callers do when an image lacks `/etc/tau/manifest.json`.

## Prerequisites

- Docker daemon running and reachable (`docker info` succeeds).
- The base image pulled locally (the demo does not pull on its own):
  ```bash
  docker pull alpine:3.21
  ```

## Run

From the `~/tau/examples` directory:

```bash
go run ./container/docker-hello/
```

## Expected output

```
hello from tau-docker-hello
manifest: <nil> (alpine carries no tau manifest)
fallback shell: /bin/sh
```

The first line is `echo` output captured from a one-shot `Exec` inside the alpine container. The second line shows that `Inspect` returns a successful `ContainerInfo` with `Manifest == nil` because vanilla alpine carries no tau manifest. The third line is the result of substituting `container.Fallback()` when the inspect-time manifest is absent — the published pattern callers should adopt when they need a non-nil `Manifest`.

## What this demonstrates

- Sub-module registration via `docker.Register()` and runtime construction via `container.Create("docker")` — the published downstream-consumer entry path.
- The eight `Runtime` methods exercised in their natural call order against a real daemon.
- The reserved `tau.managed=true` and `tau.manifest.version=<v>` labels applied by `Create`. While the container is running, inspect them with:
  ```bash
  docker ps --filter label=tau.managed=true --format '{{.Names}} {{.Labels}}'
  ```
- The nil-manifest fallback pattern: `info.Manifest` is left nil by `Inspect` when the image carries no `/etc/tau/manifest.json`, and the caller substitutes `container.Fallback()` to obtain a non-nil manifest with POSIX-shell defaults.

## Keeping the container alive for Exec

The demo passes `Cmd: []string{"sh", "-c", "trap exit TERM; sleep infinity & wait"}` to `CreateOptions`. Alpine's default CMD is `/bin/sh`, which exits immediately when started without an attached stdin/TTY — `Runtime.Exec` would then race against the container's transition to `StateExited`. The trap-and-wait pattern keeps the container running and responds cleanly to the SIGTERM that `Runtime.Stop` issues, so shutdown completes in milliseconds rather than waiting out the daemon's SIGKILL escalation.

## References

- Library docs and source: [`tailored-agentic-units/container`](https://github.com/tailored-agentic-units/container)
- Docker runtime implementation: [`tailored-agentic-units/container/docker`](https://github.com/tailored-agentic-units/container/tree/main/docker)
- Tracking issue: [container#14](https://github.com/tailored-agentic-units/container/issues/14)
