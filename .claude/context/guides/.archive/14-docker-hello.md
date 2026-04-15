# 14 - docker-hello runnable example

## Problem Context

This example is the downstream-consumer proof for [Objective #3 — Docker Runtime Implementation](https://github.com/tailored-agentic-units/container/issues/3) in the `tailored-agentic-units/container` repo. Sub-issues #11/#12/#13 implemented the `Runtime` interface end-to-end, and the Phase 1 release session has tagged `v0.1.0` (root) and `docker/v0.1.0` (Docker sub-module).

The container library now needs an example that:

- Imports the *tagged* releases from a separate Go module — no `replace` directives, no workspace refs.
- Drives the published `Runtime` API surface end-to-end against a real Docker daemon.
- Demonstrates the documented nil-manifest path explicitly (acceptance criterion #6 of [issue #14](https://github.com/tailored-agentic-units/container/issues/14)) so readers see the `container.Fallback()` substitution pattern.

This guide implements that example as the first entry in a new `container/` directory under the examples repo, mirroring the existing `orchestrate/` convention so future container-library demos (Phase 2 shell wrapper, Phase 3 image-builder, etc.) have a natural home.

## Architecture Approach

**Layout — nest under `container/`, not `cmd/`.** The `cmd/` directory is reserved for one-off CLI utilities like `prompt-agent`. Library demos live under a directory named for their library, with an index README at the top — exactly how `orchestrate/` is organized today (seven sibling demos + `orchestrate/README.md`). Adding container demos to `cmd/` and grafting on more siblings later would crowd `cmd/` and obscure that they all belong to one library. The new layout:

```
~/tau/examples/
├── cmd/prompt-agent/        # unchanged
├── orchestrate/             # unchanged
└── container/               # NEW
    ├── README.md            # AI-authored in Phase 7
    └── docker-hello/
        ├── main.go          # this guide
        └── README.md        # AI-authored in Phase 7
```

`cmd/prompt-agent/` stays where it is — a multi-protocol CLI utility, not a library demo.

**Demo style — smoke-level, not a test harness.** No assertions, no error wrapping beyond `log.Fatalf`. The example is a literate read-through of the API surface; failures should surface as a loud `Fatalf` line, equivalent to the `panic` semantics the issue describes but matching the `prompt-agent` convention used elsewhere in the repo.

**Two refinements from the issue's pseudo-code:**

1. Set `CreateOptions.Name = "tau-docker-hello"` explicitly. The issue's snippet echoes `c.Name`, but `dockerRuntime.Create` returns `Name: opts.Name` (passed through), so without an explicit `Name` the echo would print an empty string. Setting the name keeps the demo's output deterministic across runs and makes the `tau.*` labelling story easier to inspect with `docker ps --filter`.
2. After printing the nil manifest, substitute `container.Fallback()` and print `m.Shell` so the substitution pattern is visible in the output, not only in the source. This satisfies acceptance criterion #6 verbatim.

**Dependencies.** Both libraries are pulled in at their tagged versions via `go get`, and `go mod tidy` materializes the Docker SDK transitive deps. No `replace` directives.

**Out of scope for this guide (handled by AI in later phases):**

- `container/README.md`, `container/docker-hello/README.md`, and any updates to the repo-root `README.md` are documentation work — authored by the AI in Phase 7 after the developer has executed this guide and the program is verified runnable.

## Implementation

### Step 1: Add the container library dependencies

From `~/tau/examples`:

```bash
go get github.com/tailored-agentic-units/container@v0.1.0
go get github.com/tailored-agentic-units/container/docker@v0.1.0
go mod tidy
```

Note the sub-module version: pass `@v0.1.0`, not `@docker/v0.1.0`. The Git tag is `docker/v0.1.0`, but `go get` derives the prefix from the sub-module path itself and rejects an explicit `docker/` in the version string (`invalid version: ... disallowed version string`).

This adds two `require` lines to `go.mod`, materializes the Docker SDK transitive dependencies into the `// indirect` block, and updates `go.sum`. After running, verify:

- `go.mod` contains `github.com/tailored-agentic-units/container v0.1.0` and `github.com/tailored-agentic-units/container/docker v0.1.0` in the direct `require` block.
- No `replace` directives appear anywhere in the file.

### Step 2: Create the docker-hello program

**New file:** `container/docker-hello/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tailored-agentic-units/container"
	"github.com/tailored-agentic-units/container/docker"
)

func main() {
	docker.Register()

	rt, err := container.Create("docker")
	if err != nil {
		log.Fatalf("create runtime: %v", err)
	}

	ctx := context.Background()

	c, err := rt.Create(ctx, container.CreateOptions{
		Image:  "alpine:3.21",
		Name:   "tau-docker-hello",
		Labels: map[string]string{"example": "docker-hello"},
	})
	if err != nil {
		log.Fatalf("create container: %v", err)
	}

	if err := rt.Start(ctx, c.ID); err != nil {
		log.Fatalf("start: %v", err)
	}

	res, err := rt.Exec(ctx, c.ID, container.ExecOptions{
		Cmd:          []string{"echo", "hello from", c.Name},
		AttachStdout: true,
	})
	if err != nil {
		log.Fatalf("exec: %v", err)
	}
	fmt.Print(string(res.Stdout))

	info, err := rt.Inspect(ctx, c.ID)
	if err != nil {
		log.Fatalf("inspect: %v", err)
	}
	fmt.Printf("manifest: %v (alpine carries no tau manifest)\n", info.Manifest)

	// Substitute the documented fallback when Inspect returns a nil Manifest.
	m := info.Manifest
	if m == nil {
		m = container.Fallback()
	}
	fmt.Printf("fallback shell: %s\n", m.Shell)

	if err := rt.Stop(ctx, c.ID, 5*time.Second); err != nil {
		log.Fatalf("stop: %v", err)
	}
	if err := rt.Remove(ctx, c.ID, false); err != nil {
		log.Fatalf("remove: %v", err)
	}
}
```

## Remediation

### R1: Container exits before Exec — set a signal-aware long-running Cmd

**Blocker.** Running the program as written produced:

```
OCI runtime exec failed: exec failed: cannot exec in a stopped container
```

**Root cause.** `alpine:3.21`'s default CMD is `/bin/sh`. Started without an attached stdin/TTY (the `Runtime.Start` API does neither), `sh` reads EOF immediately and the container transitions straight from `StateRunning` to `StateExited`. By the time `Exec` is dispatched, the container is no longer running.

This is a property of the example image, not a defect in the runtime: the Docker sub-module faithfully starts whatever the image's CMD declares. The fix belongs in the example's `CreateOptions`.

**Fix.** Pass an explicit long-running `Cmd` that handles SIGTERM cleanly. Patch in `container/docker-hello/main.go`, inside the `rt.Create` call:

```diff
 c, err := rt.Create(ctx, container.CreateOptions{
     Image:  "alpine:3.21",
     Name:   "tau-docker-hello",
+    Cmd:    []string{"sh", "-c", "trap exit TERM; sleep infinity & wait"},
     Labels: map[string]string{"example": "docker-hello"},
 })
```

Two simpler-looking alternatives — `sleep infinity` and `tail -f /dev/null` — also keep the container alive, but each becomes PID 1 with no SIGTERM handler. Linux ignores signals to PID 1 unless the process registers a handler, so `Runtime.Stop` would have to wait the full 5 s grace period and let the daemon escalate to SIGKILL. The trap+wait pattern installs an explicit `TERM` handler, runs `sleep` in the background so the shell can `wait` on it, and exits cleanly within milliseconds when Stop fires. The Stop call's timeout becomes a ceiling, not a mandatory wait.

## Validation Criteria

- [ ] `go.mod` requires `github.com/tailored-agentic-units/container v0.1.0` and `github.com/tailored-agentic-units/container/docker v0.1.0` in the direct require block; no `replace` directives.
- [ ] `go build ./container/docker-hello/...` from `~/tau/examples` succeeds.
- [ ] `go vet ./container/docker-hello/...` clean.
- [ ] With Docker daemon running and `alpine:3.21` pulled, `go run ./container/docker-hello/` exits zero and prints exactly:
  ```
  hello from tau-docker-hello
  manifest: <nil> (alpine carries no tau manifest)
  fallback shell: /bin/sh
  ```
- [ ] `docker ps -a --filter label=tau.managed=true --filter name=tau-docker-hello` shows no rows after the run completes (Remove cleaned up).
