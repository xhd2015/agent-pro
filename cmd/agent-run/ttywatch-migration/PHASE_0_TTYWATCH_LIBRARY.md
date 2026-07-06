# Phase 0: TTYWatch Library Extensions

**Status:** pending  
**Depends on:** —  
**Blocks:** Phases 1, 2, 3, 4

## Objective

Extend `pkgs/ttywatch` so agent-run can depend on it exclusively for all PTY client and registry operations. No agent-run changes in this phase except any import-path adjustments in `script/tty-watch` adapters.

## Deliverables

### 0.1 Lift attach/watch client code into library

**Source:** `script/tty-watch/attach.go` (~770 lines)

**Target:** `pkgs/ttywatch/` (suggested files: `attach_client.go`, `observer.go`)

Export public APIs:

```go
func AttachWriter(listenAddr, sessionID, attachMode string) (detached bool, err error)
func StreamObserver(listenAddr, sessionID string, w io.Writer) error
func DialTerminal(listenAddr, sessionID, attachMode string) (*websocket.Conn, error)
```

Include: raw TTY mode, SIGWINCH resize, Ctrl-] detach, CRLF normalization, kitty pop cleanup, handshake consumption.

**Update:** `script/tty-watch/attach_cmd.go`, `watch.go`, `run.go` call library functions instead of local `attach.go` helpers. `script/tty-watch/attach.go` becomes thin re-exports or is deleted.

### 0.2 Parameterize registry location

Generalize registry so agent-run uses `AGENT_RUN_HOME/{runner}-registry/` instead of `~/.tty-watch/registry/`.

**Current:** `pkgs/ttywatch/registry.go` hardcodes `registrySubdir = "registry"` under `TTYWatchHome()`.

**Target API (conceptual):**

```go
type RegistryConfig struct {
    Home    string // AGENT_RUN_HOME or TTY_WATCH_HOME
    Subdir  string // "registry" for tty-watch; "grok-tty-registry" for agent-run
}

func RegistryDir(cfg RegistryConfig) string
func ReserveRegistrySessionID(cfg RegistryConfig) (id string, release func(), err error)
func ReserveCustomSessionID(cfg RegistryConfig, sessionID string) (release func(), err error)
func ReadRegistry(cfg RegistryConfig, sessionID string) (*RegistryEntry, error)
func WriteRegistry(cfg RegistryConfig, entry RegistryEntry) error
func RemoveRegistry(cfg RegistryConfig, sessionID string)
func RemoveRegistryIfMatch(cfg RegistryConfig, sessionID, listenAddr string, pid int)
func ListRegistryEntries(cfg RegistryConfig, prune bool) ([]RegistryEntry, error)
func WaitForRegistryEntry(cfg RegistryConfig, sessionID string, timeout time.Duration) (*RegistryEntry, error)
func TCPReachable(addr string) bool
```

Preserve backward compatibility: `TTYWatchHome()` + default `"registry"` subdir continues to work for tty-watch CLI.

Add helper for agent-run multi-runner lookup:

```go
func LookupAcrossSubdirs(home string, subdirs []string, sessionID string) (*RegistryEntry, string /*subdir*/, error)
```

### 0.3 Consolidate inject into library

Move `pkgs/ttyrunner/inject.go` → `pkgs/ttywatch/inject_input.go` (or extend existing `inject.go`).

```go
func InjectInput(listenAddr, sessionID string, input []byte) error
```

Update `script/tty-watch/send.go` and `pkgs/ttywatch/session.go` (`EphemeralSession.Send`) to use the consolidated path. `ttyrunner` re-exports temporarily (removed in Phase 5).

### 0.4 Add `SendMessage` with agent-run suffix rule

**File:** `pkgs/ttywatch/send.go`

```go
// SendMessage prepares inject mode and sends verbatim bytes.
// When suffixCR is true, appends '\r' only if message contains no '\r'.
func SendMessage(listenAddr, sessionID, message string, suffixCR bool) error
```

Implementation:

1. `PrepareSessionInjectMode(listenAddr, sessionID)`
2. `payload := message`
3. If `suffixCR && !strings.Contains(payload, "\r")` → `payload += "\r"`
4. `InjectInput(listenAddr, sessionID, []byte(payload))`

No trim, no Ctrl-U, no other transforms.

`tty-watch send` continues using `PrepareSessionInjectMode` + `InjectInput` directly (no `SendMessage`, no auto `\r`).

### 0.5 Snapshot APIs (verify / expose)

Ensure these are public and sufficient for agent-run Phase 3:

```go
func ReadSnapshot(listenAddr, sessionID string) (frame, scrollback string, cols, rows int, err error)
func SnapshotText(listenAddr, sessionID string) (string, error)
func SanitizeForPrint(data string) string
func RenderSnapshotOutput(frame, scrollback string, cols, rows int) string
```

Already exist; confirm no CLI-only wrappers block agent-run import.

## Files Touched (expected)

| Action | Path |
|--------|------|
| Create / extend | `pkgs/ttywatch/attach_client.go`, `observer.go`, `send.go`, `inject_input.go` |
| Modify | `pkgs/ttywatch/registry.go`, `inject.go`, `session.go` |
| Modify | `script/tty-watch/adapters.go`, `attach_cmd.go`, `watch.go`, `run.go`, `send.go` |
| Delete or gut | `script/tty-watch/attach.go` (after lift) |

## Acceptance Criteria

- [ ] `go build ./pkgs/ttywatch/...` passes
- [ ] `go build ./script/tty-watch/...` passes
- [ ] `script/tty-watch` attach/watch/run/send behavior unchanged (existing tty-watch doctests still pass)
- [ ] `SendMessage` unit test: verbatim + conditional `\r` cases from MIGRATE_PLAN.md
- [ ] Parameterized registry: can read/write under arbitrary `{home}/{subdir}/`
- [ ] `InjectInput` lives in `pkgs/ttywatch` only

## Subagent Prompt Template

```
Implement Phase 0 of agent-run TTY migration per:
  cmd/agent-run/ttywatch-migration/PHASE_0_TTYWATCH_LIBRARY.md

Extend pkgs/ttywatch with attach/watch client code lifted from script/tty-watch/attach.go,
parameterized registry, consolidated InjectInput, and SendMessage(suffixCR).
Update script/tty-watch to use library APIs. Do not modify cmd/agent-run yet.
Verify: go build ./pkgs/ttywatch/... ./script/tty-watch/...
```