# 14 - docker-hello runnable example

## Summary

Added `container/docker-hello/`, the first downstream-consumer demo for `github.com/tailored-agentic-units/container@v0.1.0` and `github.com/tailored-agentic-units/container/docker@v0.1.0`. The program drives the full `Runtime` lifecycle (Create → Start → Exec → Inspect → Stop → Remove) end-to-end against a local Docker daemon and demonstrates the documented `container.Fallback()` substitution for images that ship no `/etc/tau/manifest.json`. Established a new `container/` subtree under the examples repo, parallel to the existing `orchestrate/` subtree, with its own index README so future container-library demos have a natural home.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Where the demo lives | New `container/` subtree at the repo root, not `cmd/docker-hello/` | The container library will accumulate further demos in Phase 2–4 (shell wrapper, tools introspection, image-builder, alternate runtimes). `cmd/` is reserved for one-off CLI utilities like `prompt-agent`. The orchestrate library's existing `orchestrate/<demo>/` layout is the established pattern for grouping library demos. |
| `CreateOptions.Name` | Set explicitly to `"tau-docker-hello"` | The issue's pseudo-code echoes `c.Name`, but `dockerRuntime.Create` returns `Name: opts.Name` verbatim — leaving it empty would produce an empty echo string and break output determinism. The explicit name also makes the `tau.*` label story easier to surface with `docker ps --filter`. |
| Keep-alive `Cmd` | `sh -c "trap exit TERM; sleep infinity & wait"` | Alpine's default CMD is `/bin/sh`, which exits immediately without a TTY/stdin and races the demo's `Exec` call. The trap-and-wait pattern keeps the container alive AND responds to SIGTERM in milliseconds, so `Runtime.Stop`'s grace period acts as a ceiling rather than a mandatory wait. Simpler alternatives (`sleep infinity`, `tail -f /dev/null`) ignore SIGTERM as PID 1 and force the daemon to escalate to SIGKILL. Discovered as Remediation R1 during developer execution. |
| Fallback demonstration | Print `m.Shell` after substituting `container.Fallback()` for the nil `info.Manifest` | Acceptance criterion #6 requires the example to demonstrate the substitution pattern explicitly. Showing the substituted shell value in stdout makes the pattern visible at runtime, not only in the source. |
| Error handling | `log.Fatalf("step: %v", err)` at every call site | Matches the `cmd/prompt-agent/main.go` convention used elsewhere in the repo. Equivalent in user-facing behavior to the `panic` semantics the issue describes (process exits non-zero with a diagnostic line). |
| README documentation | Authored entirely in Phase 7, kept out of the implementation guide | Documentation is an AI responsibility per the workflow; the implementation guide stays focused on the source the developer executes. |

## Files Modified

**Created:**
- `container/README.md` — index for container library demos (mirrors `orchestrate/README.md`)
- `container/docker-hello/main.go` — the demo program
- `container/docker-hello/README.md` — per-demo prerequisites, run command, expected output, and the keep-alive `Cmd` rationale
- `.claude/context/guides/.archive/14-docker-hello.md` — archived implementation guide (with R1 remediation)
- `.claude/context/sessions/14-docker-hello.md` — this summary

**Modified:**
- `go.mod`, `go.sum` — added `tau/container v0.1.0` and `tau/container/docker v0.1.0` to the direct require block; transitive Docker SDK deps materialized via `go mod tidy`
- `README.md` (repo root) — added `container/` to the Structure block, a `container examples` bullet to Sub-READMEs, and two rows to the Dependencies table

## Patterns Established

- **`container/` subtree convention.** Future container-library demos live as siblings of `docker-hello/`, indexed in `container/README.md`. Mirrors the existing `orchestrate/` subtree treatment so the repo has one shape for "library demos" regardless of which library.
- **Keep-alive idiom for runtime demos.** `Cmd: []string{"sh", "-c", "trap exit TERM; sleep infinity & wait"}` is the recommended pattern when an Exec-style demo needs the container to outlive the image's default CMD without delaying Stop. Worth reusing in any future container demo that requires a running container the runtime drives interactively.
- **Documentation deferred to Phase 7.** Implementation guides under `.claude/context/guides/` contain only the source the developer types — no READMEs, no doc comments, no PM updates. All documentation work is AI-handled after the developer has signaled the program runs.

## Validation Results

- `go vet ./container/docker-hello/...` — clean.
- `go build ./container/docker-hello/...` — succeeds.
- `go.mod` — `tau/container v0.1.0` and `tau/container/docker v0.1.0` present in the direct `require` block; no `replace` directives.
- `go run ./container/docker-hello/` against a local Docker daemon with `alpine:3.21` pulled — exits zero, prints exactly:
  ```
  hello from tau-docker-hello
  manifest: <nil> (alpine carries no tau manifest)
  fallback shell: /bin/sh
  ```
- `docker ps -a --filter label=tau.managed=true --filter name=tau-docker-hello` — empty after the run (Remove cleaned up).

## Follow-ups

- `~/tau/container/_project/objective.md` row for sub-issue D needs to flip from "Todo, blocked on release" to "Done" after the examples PR merges and `container#14` closes.
- The examples repo has no `.gitignore` — `go build ./pkg/...` on a `main` package writes the binary to CWD, which produced a stray `docker-hello` ELF in this session that had to be removed manually before commit. Worth bootstrapping a Go-flavored `.gitignore` in a follow-up.
