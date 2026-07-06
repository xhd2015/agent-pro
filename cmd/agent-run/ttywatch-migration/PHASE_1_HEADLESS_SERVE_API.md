# Phase 1: Headless Detached Serve API

**Status:** pending  
**Depends on:** Phase 0  
**Blocks:** Phase 2

## Objective

Expose a programmatic headless + detached `__serve__` API in `pkgs/ttywatch` matching `tty-watch run --headless` behavior. agent-run `run` will use this instead of in-process `groktty.Run`.

## Background

`tty-watch run` today:

1. Reserves session ID in registry
2. Re-execs self with `__serve__` token + session ID + command argv
3. Waits for registry entry
4. `--headless`: prints `session-id`, waits on serve child exit
5. Default (non-headless): attaches as writer — **not used by agent-run**

agent-run always runs headless: stdout is reserved for agent events, not TTY layout.

## Deliverables

### 1.1 `HeadlessRun` API

**File:** `pkgs/ttywatch/headless.go` (suggested)

```go
type HeadlessRunOptions struct {
    Home           string
    RegistrySubdir string        // e.g. "grok-tty-registry"
    SessionID      string        // empty → auto-reserve session-N
    Command        []string
    BinaryPath     string        // re-exec target for __serve__ (os.Args[0])
    Cwd            string
    ExtraPaths     []string
    KeepAlive      bool          // do not wait for child exit; return after registry ready
}

type HeadlessRunResult struct {
    SessionID  string
    Entry      *RegistryEntry
    Cmd        *exec.Cmd       // serve child process
    Wait       func() error    // wait for serve child exit
}

func HeadlessRun(ctx context.Context, opts HeadlessRunOptions) (*HeadlessRunResult, error)
```

Behavior:

1. Reserve session ID via parameterized registry (Phase 0)
2. Build argv: `[binaryPath, ServeSubcommand(command), sessionID, ...command]`
3. Start detached child (`Setsid: true`, stdin/stdout/stderr nil)
4. `WaitForRegistryEntry` with timeout (15s)
5. Return result; caller decides whether to `Wait()` or detach further

### 1.2 SIGINT / graceful interrupt forwarding

Port headless interrupt logic from `script/tty-watch/run.go`:

- `forwardHeadlessInterrupt` — kill process group or inject `\x03` via `PrepareSessionInjectMode` + `InjectInput`
- `waitHeadless` — SIGINT grace window, "waiting for program to exit..." stderr message

Expose optionally:

```go
func WaitHeadless(ctx context.Context, result *HeadlessRunResult, command []string) error
```

Or keep internal to `HeadlessRun` when `KeepAlive == false`.

### 1.3 `__serve__` entry unchanged

`ttywatch.ServeSession` + `ttywatch.IsServeSubcommand` already exist. Ensure `HeadlessRun` uses the same serve token generation (`ServeSubcommand(command)`).

### 1.4 `EphemeralSession` alignment

Update `pkgs/ttywatch/session.go` `StartDetached` to use parameterized registry from Phase 0 (if not already). `EphemeralSession` becomes the in-process alternative; `HeadlessRun` is the detached default for agent-run.

## Files Touched (expected)

| Action | Path |
|--------|------|
| Create | `pkgs/ttywatch/headless.go` |
| Modify | `pkgs/ttywatch/session.go`, `server.go`, `naming.go` |
| Modify | `script/tty-watch/run.go` — delegate to `HeadlessRun` where possible |

## Acceptance Criteria

- [ ] `go build ./pkgs/ttywatch/...` passes
- [ ] `HeadlessRun` spawns detached serve child and returns registry entry
- [ ] Session survives parent exit (detach semantics)
- [ ] SIGINT forwarding matches tty-watch headless behavior
- [ ] `script/tty-watch run --headless` can delegate to `HeadlessRun` without behavior change
- [ ] Existing tty-watch headless doctests still pass

## Subagent Prompt Template

```
Implement Phase 1 of agent-run TTY migration per:
  cmd/agent-run/ttywatch-migration/PHASE_1_HEADLESS_SERVE_API.md

Add HeadlessRun API to pkgs/ttywatch for detached __serve__ child with headless wait
and SIGINT forwarding. Refactor script/tty-watch/run.go to use it where possible.
Depends on Phase 0 (parameterized registry, inject). Do not modify cmd/agent-run yet.
```