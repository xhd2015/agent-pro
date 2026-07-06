# agent-run TTY → pkgs/ttywatch Migration Plan

Migrate all agent-run TTY operations to `pkgs/ttywatch`, the shared PTY session library already used by `tty-watch`. This is a **complete refactor**: duplicate server, registry, attach, send, and WebSocket code in agent-run is deleted, not shimmed.

## Goals

| Goal | Decision |
|------|----------|
| PTY engine | `pkgs/ttywatch` only (`ServeSession`, registry, attach, watch, snapshot, send) |
| Process model | Detached `__serve__` child (tty-watch pattern) |
| `agent-run run` | Always headless — stdout shows agent session events, not TTY layout |
| Send semantics | `PrepareSessionInjectMode` + verbatim message; append `\r` only if message contains no `\r` |
| New CLI commands | `snapshot`, `watch` added to agent-run (top-level + `tty` subcommand) |
| Legacy code | Delete all agent-run / groktty PTY code not shared via the core library |
| Tests | Migrate later (Phase 7) |
| Web terminal | Migrate in Phase 5 |

## Target Architecture

```
cmd/agent-run
  ├─ run (headless, detached serve)
  ├─ attach / send / snapshot / watch
  ├─ tty {status,attach,send,snapshot,watch}
  └─ web terminal proxy
       │
       ▼
pkgs/agenttty (slim agent adapter — argv, banner, tail, writable, tty.json)
       │
       ▼
pkgs/ttywatch (sole PTY core)
  ├─ ServeSession / detached __serve__
  ├─ registry (parameterized home + subdir)
  ├─ AttachWriter / StreamObserver
  ├─ SnapshotText / ReadSnapshot
  └─ PrepareSessionInjectMode + InjectInput + SendMessage
       │
       ▼
ptywrap
```

**Principle:** If `pkgs/ttywatch` can do it, agent-run must not reimplement it.

## Phase Overview

| Phase | File | Summary | Status |
|-------|------|---------|--------|
| 0 | [PHASE_0_TTYWATCH_LIBRARY.md](./PHASE_0_TTYWATCH_LIBRARY.md) | Extend `pkgs/ttywatch`: lift attach/watch, parameterize registry, consolidate inject, `SendMessage` | **completed** |
| 1 | [PHASE_1_HEADLESS_SERVE_API.md](./PHASE_1_HEADLESS_SERVE_API.md) | Headless detached `__serve__` run API in library | **completed** |
| 2 | [PHASE_2_AGENT_RUN_RUN.md](./PHASE_2_AGENT_RUN_RUN.md) | Replace `agent-run run` TTY path with detached serve + agent adapter | **completed** |
| 3 | [PHASE_3_CLI_COMMANDS.md](./PHASE_3_CLI_COMMANDS.md) | `attach`, `send`, `snapshot`, `watch`, `tty status` on ttywatch | **completed** |
| 4 | [PHASE_4_WEB_TERMINAL.md](./PHASE_4_WEB_TERMINAL.md) | Migrate `web_terminal.go` to ttywatch APIs | **completed** |
| 5 | [PHASE_5_DELETE_LEGACY.md](./PHASE_5_DELETE_LEGACY.md) | Delete groktty server/registry, duplicate inject/attach/WS code | **completed** |
| 6 | [PHASE_6_TEST_MIGRATION.md](./PHASE_6_TEST_MIGRATION.md) | Update agent-run doctests (deferred) | **deferred** |

Track live status in [PHASE_STATUS.md](./PHASE_STATUS.md).

## Send Semantics (locked)

Applies to `agent-run send`, `agent-run tty send`, and web follow-up send.

| Rule | Detail |
|------|--------|
| Base | `PrepareSessionInjectMode` then verbatim inject |
| Suffix | Append `\r` only when message contains no `\r` anywhere |
| Forbidden | `strings.TrimSpace`, `\x15` (Ctrl-U), LF coercion, other transforms |
| tty-watch CLI | Unchanged — verbatim, no auto `\r` (uses `InjectInput` directly) |

Examples:

| Input | Injected bytes |
|-------|----------------|
| `"follow up"` | `"follow up\r"` |
| `"follow up\r"` | `"follow up\r"` |
| `"line1\nline2"` | `"line1\nline2\r"` |
| `"already\r\n"` | `"already\r\n"` |

## New agent-run Commands

Mirror tty-watch surface area:

```
agent-run snapshot <session-id>    # sanitized scrollback to stdout
agent-run watch <session-id>       # readonly observe (stdout)
agent-run tty snapshot <session-id>
agent-run tty watch <session-id>
```

Registry lookup uses `AGENT_RUN_HOME/{runner}-registry/` via parameterized ttywatch registry (multi-runner search preserved).

## PR / Subagent Execution Order

```
Phase 0 ──► Phase 1 ──► Phase 2
              │
              └──► Phase 3 ──► Phase 4 ──► Phase 5
                                    │
                                    └──► Phase 6 (later)
```

- **Phase 0** must complete before all others.
- **Phase 1** must complete before Phase 2.
- **Phase 3** requires Phase 0; can start after Phase 0 (parallel with Phase 1/2 if careful).
- **Phase 4** requires Phase 0 + Phase 3.
- **Phase 5** requires Phases 2, 3, 4.
- **Phase 6** is deferred until Phase 5 lands.

## Key Behavioral Changes

| Area | Before | After |
|------|--------|-------|
| `run` process | In-process ptywrap in agent-run | Detached `__serve__` child |
| `run` stdout | Agent events (no layout attach) | Explicitly headless — events only |
| `send` payload | `\x15` + trim + `\r` | Verbatim + conditional `\r` |
| `attach` | Basic `ptyclient` | Full ttywatch attach (detach, resize) |
| `snapshot` / `watch` | Not available | New commands via ttywatch |
| Registry | groktty + ttywatch duplicate | ttywatch only (parameterized paths) |
| Web follow-up | Ctrl-U WS inject | `SendMessage` via HTTP inject |
| Server API | No `prepare-inject` on groktty servers | Always via `ttywatch.ServeSession` |

## Out of Scope

- `agentstorage` session/event persistence model
- Non-TTY runners (`codex`, `opencode`, `fake-codex`, etc.)
- `agent-run sessions` / `status` commands (except tty resolution paths they touch)
- Merging `AGENT_RUN_HOME` and `TTY_WATCH_HOME` into one directory

## References

- Core library: `pkgs/ttywatch/`
- tty-watch CLI (reference impl): `script/tty-watch/`
- Current agent-run TTY code: `cmd/agent-run/tty_cmd.go`, `cmd/agent-run/web_terminal.go`
- Legacy to delete: `pkgs/groktty/runner.go`, `pkgs/groktty/registry.go`, `pkgs/ttyrunner/run.go`, `pkgs/ttyrunner/inject.go`, `pkgs/ttyrunner/attach_client.go`, `pkgs/ttyrunner/scrollback.go`