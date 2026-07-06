# Phase 5: Delete Legacy Code

**Status:** pending  
**Depends on:** Phases 2, 3, 4  
**Blocks:** Phase 6

## Objective

Remove all agent-run / groktty / ttyrunner code that duplicated `pkgs/ttywatch`. After this phase, there is a single PTY stack.

## Delete Entirely

| Path | Reason |
|------|--------|
| `pkgs/groktty/runner.go` | In-process server replaced by `ttywatch.ServeSession` / `HeadlessRun` |
| `pkgs/groktty/registry.go` | Replaced by parameterized `ttywatch` registry |
| `pkgs/ttyrunner/run.go` | Replaced by `agenttty.RunHeadless` |
| `pkgs/ttyrunner/inject.go` | Moved to `pkgs/ttywatch` in Phase 0 |
| `pkgs/ttyrunner/attach_client.go` | Replaced by `pkgs/ttywatch` attach client |
| `pkgs/ttyrunner/scrollback.go` | Replaced by `pkgs/ttywatch` snapshot APIs |
| `script/tty-watch/attach.go` | Lifted to `pkgs/ttywatch` in Phase 0 (if not already deleted) |

## Gut `pkgs/groktty`

After Phase 2 extraction to `pkgs/agenttty`, delete remaining groktty files that only served the old server path.

**Keep temporarily** (if still referenced by non-TTY code):

| File | Check before delete |
|------|---------------------|
| `pkgs/groktty/ansi.go` | Used by agenttty capture? → move to agenttty |
| `pkgs/groktty/env.go` | Used by agenttty? → move |
| `pkgs/groktty/path_escape.go` | Used by agenttty argv? → move |

**Target:** `pkgs/groktty/` package deleted entirely; all imports updated to `pkgs/agenttty` or `pkgs/ttywatch`.

Run before deleting:

```sh
rg -l "pkgs/groktty" --glob '*.go'
```

## Shrink `pkgs/ttyrunner`

After migration, `pkgs/ttyrunner` should not exist as a separate package OR should be reduced to nothing.

**Preferred:** Delete `pkgs/ttyrunner/` entirely; all symbols live in:

- `pkgs/ttywatch` — generic PTY
- `pkgs/agenttty` — agent providers, writable, resolution, tty.json

### Migration map

| ttyrunner symbol | New location |
|------------------|--------------|
| `Provider`, `Register`, `Get`, `IDs`, `IsTTYRunner` | `pkgs/agenttty` |
| `BuildArgv` hooks | `pkgs/agenttty` |
| `CheckWritable`, `DetectScreenStatus` | `pkgs/agenttty` |
| `ResolveByTerminalID`, `ResolveByAgentSession`, `LookupSession` | `pkgs/agenttty` |
| `WriteTTYJSON`, `TTYSnapshot` | `pkgs/agenttty` |
| `InjectInput` | `pkgs/ttywatch` (Phase 0) |
| `FetchScrollbackSnapshot` | `pkgs/ttywatch.SnapshotText` |
| `Run`, `RunStubTTYMain` | `pkgs/agenttty` |
| `WSAttachClient` | `pkgs/ttywatch` |

Update all imports:

```sh
rg -l "pkgs/ttyrunner" --glob '*.go'
rg -l "pkgs/groktty" --glob '*.go'
```

## Clean `cmd/agent-run`

Verify no dead code remains in:

- `cmd/agent-run/tty_cmd.go` — no `injectViaWebSocketSnapshot`, no ptyclient
- `cmd/agent-run/web_terminal.go` — no `\x15`, no raw inject WS
- `cmd/agent-run/main.go` — `__stub-tty` points to agenttty

## `script/tty-watch`

Ensure adapters only re-export or call `pkgs/ttywatch` — no duplicated attach logic.

## Acceptance Criteria

- [ ] `rg "pkgs/groktty" --glob '*.go'` returns no matches (or only deleted package)
- [ ] `rg "pkgs/ttyrunner" --glob '*.go'` returns no matches (or only deleted package)
- [ ] `rg '\\x15' cmd/agent-run/` returns no matches
- [ ] `rg 'ptyclient' cmd/agent-run/` returns no matches
- [ ] `go build ./...` passes
- [ ] `go test ./pkgs/...` passes (non-doctest unit tests)

## Subagent Prompt Template

```
Implement Phase 5 of agent-run TTY migration per:
  cmd/agent-run/ttywatch-migration/PHASE_5_DELETE_LEGACY.md

Delete pkgs/groktty server/registry and pkgs/ttyrunner after migrating all
imports to pkgs/agenttty and pkgs/ttywatch. Verify with rg and go build ./...
Depends on Phases 2, 3, 4 being complete.
```