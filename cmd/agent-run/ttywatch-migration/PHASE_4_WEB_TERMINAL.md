# Phase 4: Web Terminal Migration

**Status:** pending  
**Depends on:** Phase 0, Phase 3  
**Blocks:** Phase 5

## Objective

Migrate `cmd/agent-run/web_terminal.go` from hand-rolled WebSocket inject/attach/snapshot code to `pkgs/ttywatch` APIs. Align web follow-up send with CLI send semantics.

## Current Problems

`web_terminal.go` duplicates logic that Phase 0–3 consolidate:

| Function | Issue |
|----------|-------|
| `buildTerminalPromptInput` | Uses `\x15` Ctrl-U prefix — violates send semantics |
| `sendPromptToLiveTerminal` | Raw WebSocket binary write, no `prepare-inject` |
| `terminalUpstreamWebSocketURL` | Hand-built WS URL |
| `captureTerminalAssistantText` | Custom snapshot read over WS |
| `resolveTerminal` | Uses `ttyrunner.ResolveByAgentSession` (ok; migrate to agenttty) |

## Target Implementation

### 4.1 Session resolution

Replace `ttyrunner` imports with `agenttty` + `ttywatch`:

```go
func resolveTerminal(store agentstorage.Store, runner, sessionID string) (*terminalResolution, bool) {
    if !agenttty.IsTTYRunner(runner) { return nil, false }
    // agenttty.ResolveByAgentSession → registry entry via ttywatch
}
```

### 4.2 WebSocket attach proxy

Browser ↔ agent-run ↔ ptywrap upstream.

Keep the browser proxy pattern but use ttywatch for upstream URL construction:

```go
upstreamURL, err := ttywatch.TerminalWebSocketURL(entry.ListenAddr, terminalSessionID, attachMode)
```

If `TerminalWebSocketURL` is not exported after Phase 0, export it from `pkgs/ttywatch/attach_client.go`.

Proxy logic (`proxyWebSocketMessages`, `proxyInitialTerminalMessages`) may remain in `web_terminal.go` — it is HTTP-handler glue, not PTY client logic.

### 4.3 Follow-up send (web)

**Delete:** `buildTerminalPromptInput`, raw WS inject in `sendPromptToLiveTerminal`.

**Use:**

```go
func sendPromptToLiveTerminal(store, runner, sessionID, prompt string) bool {
    resolved, ok := resolveTerminal(store, runner, sessionID)
    if !ok || resolved.Entry == nil { return false }

    provider, _ := agenttty.Get(resolved.Runner)
    writable := agenttty.WaitUntilWritable(provider, resolved.Entry.ListenAddr, resolved.TerminalSessionID, 10*time.Second)
    if !writable.Ready { return false }

    err := ttywatch.SendMessage(resolved.Entry.ListenAddr, resolved.TerminalSessionID, prompt, suffixCR=true)
    if err != nil { return true } // or false per existing contract

    // Optional: capture response via SnapshotText polling instead of WS snapshot
    text := captureAssistantFromSnapshot(resolved, prompt, runner)
    // append AgentEvents, ActionDone, update session status
    return true
}
```

### 4.4 Response capture

**Delete:** `captureTerminalAssistantText` WS loop (or reduce to fallback).

**Prefer:** Poll `ttywatch.SnapshotText` + `agenttty` scrollback extraction (same as CLI fallback path in Phase 2).

```go
func captureAssistantFromSnapshot(resolved *terminalResolution, prompt, runner string) string {
    // poll SnapshotText with idle timeout
    // extract via agenttty/capture helpers
}
```

### 4.5 Terminal status endpoint

`handleTerminalStatus`: use `ttywatch.TCPReachable` + registry entry from agenttty resolution. No behavior change for JSON shape.

### 4.6 Writable / sendable state for web UI

If web handlers expose sendable state, use `agenttty.WaitUntilWritable` / `CheckWritable` on `SnapshotText` — same as `tty status` (Phase 3).

## Files Touched (expected)

| Action | Path |
|--------|------|
| Modify | `cmd/agent-run/web_terminal.go` |
| Modify | `cmd/agent-run/web_handlers.go` (if terminal routes reference old helpers) |
| Modify | `cmd/agent-run/web.go` (if needed) |

## Acceptance Criteria

- [ ] `go build ./cmd/agent-run/...` passes
- [ ] Web terminal WebSocket proxy still connects browser to live session
- [ ] Web follow-up send uses `SendMessage(suffixCR=true)` — no Ctrl-U
- [ ] No `buildTerminalPromptInput` or `\x15` in web_terminal.go
- [ ] No direct `ttyrunner.InjectInput` in web_terminal.go
- [ ] Response capture uses `SnapshotText` (WS snapshot fallback removed or minimal)
- [ ] `resolveTerminal` uses agenttty + ttywatch registry

## Subagent Prompt Template

```
Implement Phase 4 of agent-run TTY migration per:
  cmd/agent-run/ttywatch-migration/PHASE_4_WEB_TERMINAL.md

Migrate cmd/agent-run/web_terminal.go to pkgs/ttywatch SendMessage, SnapshotText,
and attach WS URL helpers. Remove Ctrl-U inject and raw WS send paths.
Depends on Phases 0 and 3.
```