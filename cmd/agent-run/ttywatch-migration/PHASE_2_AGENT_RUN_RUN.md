# Phase 2: agent-run run (Headless TTY Path)

**Status:** pending  
**Depends on:** Phase 0, Phase 1  
**Blocks:** Phase 5

## Objective

Replace `agent-run run` for TTY runners (`grok-tty`, `codex-tty`, `stub-tty`) with detached `__serve__` via `ttywatch.HeadlessRun`. stdout streams agent events only; no PTY layout attach.

## Current Path (to replace)

```
agentui.Run
  → ttyrunner.Run
    → groktty.Run (in-process ptywrap, direct mgr.WriteInput, no prepare-inject)
```

## Target Path

```
agentui.Run
  → agenttty.RunHeadless (new slim package)
    → ttywatch.HeadlessRun (detached serve)
    → agenttty: banner wait, SendMessage(prompt), event tail, scrollback fallback
    → agentstorage: events, tty.json dual-write
```

## Deliverables

### 2.1 Create `pkgs/agenttty` (agent adapter)

Extract agent-specific logic from `pkgs/groktty` that is **not** PTY server infrastructure:

| Module | Source | Purpose |
|--------|--------|---------|
| `argv.go` | `groktty/command.go` | `BuildGrokCommandArgv`, `BuildCodexCommandArgv` |
| `config_home.go` | `groktty/config_home.go` | `ResolveAgentRunnerConfigHome`, env prepend |
| `banner.go` | `groktty/runner.go` | `waitForBannerConfig` (poll via `SnapshotText`) |
| `tail_grok.go` | `groktty/tail.go` | Grok updates.jsonl tailing |
| `tail_codex.go` | `groktty/codex_tail.go` | Codex transcript tailing |
| `capture.go` | `groktty/capture.go` | Scrollback text extraction fallback |
| `writable.go` | `pkgs/ttyrunner/writable.go` | Grok/codex writable heuristics |
| `provider.go` | `pkgs/ttyrunner/registry.go` | Provider registry (IDs, BuildArgv, CheckWritable) |
| `run.go` | new | `RunHeadless(ctx, RunOptions) (runnerSessionID, terminalSessionID string, err error)` |

**Do not copy** into agenttty: `ptywrap.NewManager`, `srv.Serve`, groktty registry functions.

### 2.2 `agenttty.RunHeadless` flow

1. Resolve provider by `RunnerID` (`grok-tty`, `codex-tty`, etc.)
2. Build argv via `provider.BuildArgv`
3. Prepend agent config env (`GROK_HOME` / `CODEX_HOME`)
4. Call `ttywatch.HeadlessRun` with:
   - `Home` = `AGENT_RUN_HOME`
   - `RegistrySubdir` = `provider.RegistryDir` (e.g. `grok-tty-registry`)
   - `BinaryPath` = agent-run binary (re-exec for `__serve__`)
   - `Command` = argv
   - `KeepAlive` = `opts.KeepTerminalAlive`
5. Print `{runner-id}: {session-id}` on stderr (existing convention)
6. Dual-write `tty.json` via callback on terminal session ID
7. Poll `SnapshotText` / writable until banner ready
8. Send initial prompt via `ttywatch.SendMessage(listenAddr, sessionID, prompt, suffixCR=true)`
9. Start event tail goroutines (grok updates / codex transcript)
10. If not `KeepTerminalAlive`: `Wait()` on serve child; else return with live session
11. Emit `AgentEvent`s; fallback to scrollback capture if tail misses
12. Update `agentstorage` session status

### 2.3 Update `pkgs/agentui/run.go`

Replace `ttyrunner.Run` call with `agenttty.RunHeadless`:

```go
// streamRunner TTY branch
if agenttty.IsTTYRunner(runner) {
    return agenttty.RunHeadless(ctx, ...)
}
```

### 2.4 `__stub-tty` / doctest harness

`cmd/agent-run/main.go` has `__stub-tty` → `ttyrunner.RunStubTTYMain()`. Migrate stub to `agenttty` or keep minimal stub entry in agent-run that uses `ttywatch.HeadlessRun` with fake argv.

Ensure `AGENT_RUN_ENABLE_STUB_TTY=1` still registers `stub-tty` provider.

### 2.5 Headless-only guarantee

- No attach to stdout/stderr for PTY layout during `agent-run run`
- Agent events on stdout (`--json` NDJSON or human-readable formatting)
- Session diagnostics on stderr only

## Files Touched (expected)

| Action | Path |
|--------|------|
| Create | `pkgs/agenttty/*.go` |
| Modify | `pkgs/agentui/run.go` |
| Modify | `cmd/agent-run/main.go` (stub entry) |
| Modify | `cmd/agent-run/run_cmd.go` (if needed for binary path) |
| Deprecate (Phase 5 delete) | `pkgs/ttyrunner/run.go`, `pkgs/groktty/runner.go` |

## Acceptance Criteria

- [ ] `go build ./...` passes
- [ ] `agent-run run --agent-runner grok-tty "prompt"` starts detached serve, prints session-id on stderr
- [ ] stdout shows agent events, not raw TTY layout
- [ ] `--keep-tty` leaves session alive after run
- [ ] `tty.json` dual-write still works
- [ ] Stub TTY runner still works with `AGENT_RUN_ENABLE_STUB_TTY=1`
- [ ] No import of `groktty.Run` from agent-run path

## Subagent Prompt Template

```
Implement Phase 2 of agent-run TTY migration per:
  cmd/agent-run/ttywatch-migration/PHASE_2_AGENT_RUN_RUN.md

Create pkgs/agenttty with headless TTY run using ttywatch.HeadlessRun.
Update agentui.Run to use agenttty instead of ttyrunner.Run/groktty.Run.
Always headless — no PTY layout on stdout. Depends on Phases 0 and 1.
```