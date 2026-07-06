# Phase 3: CLI Commands (attach, send, snapshot, watch, tty status)

**Status:** pending  
**Depends on:** Phase 0  
**Blocks:** Phase 4, Phase 5

## Objective

Rewire all agent-run TTY CLI commands to `pkgs/ttywatch`. Add `snapshot` and `watch` commands. Delete custom attach/send/scrollback code from `tty_cmd.go`.

## Command Surface (final)

### Top-level (shortcuts)

```
agent-run attach <session-id>
agent-run send <session-id> <message...>
agent-run snapshot <session-id>
agent-run watch <session-id>
```

### tty subcommand

```
agent-run tty status [--json] <session-id>
agent-run tty attach <session-id>
agent-run tty send <session-id> <message...>
agent-run tty snapshot <session-id>
agent-run tty watch <session-id>
```

Update help text in `cmd/agent-run/main.go` and `cmd/agent-run/tty_cmd.go`.

## Session Resolution

All commands resolve terminal session by registry ID across agent-run home:

```go
home := store.Home() // AGENT_RUN_HOME
subdirs := agenttty.ProviderRegistrySubdirs() // ["grok-tty-registry", "codex-tty-registry", ...]
entry, subdir, err := ttywatch.LookupAcrossSubdirs(home, subdirs, sessionID)
```

Stale unreachable entries: prune per ttywatch `RemoveRegistryIfMatch`.

Link to agent session (for `tty status` metadata): `agenttty.ResolveAgentSession(home, sessionID)` using `tty.json` + `agentstorage` meta.

## Per-Command Implementation

### attach / tty attach

**Delete:** `ptyclient.Attach` usage.

**Use:**

```go
entry, _, err := resolveRegistryEntry(sessionID)
_, err = ttywatch.AttachWriter(entry.ListenAddr, sessionID, "attach")
```

### send / tty send

**Delete:**

- `\x15` prefix (`[]byte("\x15" + ...)`)
- `strings.TrimSpace` on message
- `injectViaWebSocketSnapshot` fallback
- `readWSResponseAfterInject`

**Use:**

```go
message := strings.Join(remaining[1:], " ") // preserve whitespace within args; no TrimSpace
provider, _ := agenttty.Get(runnerID)
writable := agenttty.WaitUntilWritable(provider, entry.ListenAddr, sessionID, 10*time.Second)
if !writable.Ready { return timeout error }
err = ttywatch.SendMessage(entry.ListenAddr, sessionID, message, suffixCR=true)
```

Optionally append `AgentEvent` to store (existing behavior) without WS response capture.

### snapshot / tty snapshot (NEW)

**Reference:** `script/tty-watch/snapshot.go`

```go
entry, _, err := resolveRegistryEntry(sessionID)
text, err := ttywatch.SnapshotText(entry.ListenAddr, sessionID)
if text != "" { fmt.Println(text) }
```

Silent success on empty snapshot (match tty-watch).

### watch / tty watch (NEW)

**Reference:** `script/tty-watch/watch.go`

```go
entry, _, err := resolveRegistryEntry(sessionID)
return ttywatch.StreamObserver(entry.ListenAddr, sessionID, os.Stdout)
```

Readonly observe — no stdin forwarding.

### tty status

**Delete:** `ttyrunner.FetchScrollbackSnapshot` ad-hoc usage in `tty_cmd.go`.

**Use:**

- Registry + TCP reachability from ttywatch
- `ttywatch.SnapshotText` for live screen content
- `agenttty` provider `CheckWritable` / `DetectScreenStatus` on snapshot text
- `agentstorage` for `session_file_path`

## File Structure

Refactor `cmd/agent-run/tty_cmd.go`:

| Option A | Option B |
|----------|----------|
| Keep single file, replace internals | Split: `tty_attach.go`, `tty_send.go`, `tty_snapshot.go`, `tty_watch.go`, `tty_status.go` |

Either is fine; prefer split if `tty_cmd.go` exceeds ~300 lines after refactor.

Add top-level command files:

```
cmd/agent-run/snapshot_cmd.go  → runSnapshot → runTtySnapshot
cmd/agent-run/watch_cmd.go     → runWatch → runTtyWatch
```

Wire in `cmd/agent-run/main.go` switch.

## Files Touched (expected)

| Action | Path |
|--------|------|
| Modify | `cmd/agent-run/main.go` |
| Modify / split | `cmd/agent-run/tty_cmd.go` |
| Create | `cmd/agent-run/snapshot_cmd.go`, `watch_cmd.go` |
| Modify | `cmd/agent-run/attach_cmd.go`, `send_cmd.go` (unchanged delegation) |
| Delete from tty_cmd.go | `injectViaWebSocketSnapshot`, `readWSResponseAfterInject`, `injectTTYInput`, ptyclient import |

## Acceptance Criteria

- [ ] `go build ./cmd/agent-run/...` passes
- [ ] `agent-run attach <id>` uses ttywatch attach (detach/resize work in real TTY)
- [ ] `agent-run send <id> msg` uses `SendMessage` — no Ctrl-U, conditional `\r` only
- [ ] `agent-run snapshot <id>` prints sanitized scrollback
- [ ] `agent-run watch <id>` streams readonly output
- [ ] `agent-run tty status` uses snapshot + agenttty writable heuristics
- [ ] Help text lists new commands
- [ ] No `ptyclient` import in agent-run TTY commands
- [ ] No `ttyrunner.InjectInput` in cmd/agent-run (use ttywatch)

## Subagent Prompt Template

```
Implement Phase 3 of agent-run TTY migration per:
  cmd/agent-run/ttywatch-migration/PHASE_3_CLI_COMMANDS.md

Rewire attach/send/status to pkgs/ttywatch. Add snapshot and watch commands
(top-level + tty subcommand). Use SendMessage with suffixCR=true for send.
Delete WS inject fallback and ptyclient usage. Depends on Phase 0.
```