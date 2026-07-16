# Supporting a New Agent Runner

This guide documents the full process of adding a new agent runner to the
agent-pro system, from initial registration through E2E testing. There are two
paths: **CLI runners** (spawned as subprocesses, speaking a wire protocol) and
**TTY runners** (running inside a PTY, observed via scrollback snapshots).

---

## Choose Your Path

| Trait | CLI Runner | TTY Runner |
|-------|-----------|------------|
| How it runs | `os/exec` → reads stdout JSON lines | PTY (pseudo-terminal) → observes scrollback |
| Examples | `opencode`, `codex` (non-tty), `crush`, `pi`, `fake-codex` | `grok-tty`, `codex-tty`, `commandcode-tty` |
| Interaction | Non-interactive; prompt as CLI arg; exits after response | Interactive TUI; inject prompt via `SendMessage` |
| Output capture | Parse JSON events from stdout line by line | Snapshot PTY scrollback, extract text, filter TUI chrome |
| Key file | `agent/cli/<runner>/<runner>.go` | `pkgs/agenttty/<runner>.go` |

**When to pick TTY**: the agent's CLI has no non-interactive JSON mode, or its
primary UX is a rich terminal UI you want to observe and attach to.

---

## Phase 1: CLI Agent Foundation (shared by both paths)

### 1.1 Register the Runner ID

File: `agent/cli/registry/settings_cli.go`

```go
AgentRunner<Runner> AgentRunnerID = "<runner>"
```

For TTY runners: register the ID as a constant even if you don't use the CLI
agent path — `runner_validate.go` needs it, and `agenttty.TTYRunner` checks
`agenttty.IsTTYRunner()` separately.

### 1.2 Runner Validation

File: `cmd/cmd/agent-run/runner_validate.go`

Add to the switch for non-TTY runners:

```go
case registry.AgentRunner<Runner>:
    return nil
```

TTY runners pass validation via `agenttty.IsTTYRunner()` — no switch case needed.

---

## Phase 2A: CLI Runner Implementation

### 2A.1 Define Wire Format Types

Create `agent/event/<runner>_types/types.go`:

- Event envelope with type + payload discriminator
- Message parts: text, reasoning, tool_call, tool_result, finish
- Run complete event with session_id

### 2A.2 Write Converters

Create `agent/event/<runner>_types/convert.go`:

- `From<Runner>(events []Event, sessionID string) []types.AgentEvent`
- `To<Runner>(events []types.AgentEvent, sessionID string) []Event`

### 2A.3 Implement the Agent

Create `agent/cli/<runner>/<runner>.go` implementing `registry.Agent`:

```go
type <Runner>Agent struct {
    AgentPath    string
    SettingsPath string
    Workspace    string
    Env          *exec.Env
    LastSessionID string
}

func (a *<Runner>Agent) Ask(ctx context.Context, question string,
    opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
    // 1. Resolve binary path
    // 2. Build command args
    // 3. Spawn process with a.Env.CommandContext
    // 4. Read stdout with bufio.Scanner
    // 5. Parse JSON events, call onDelta for text
    // 6. Store session ID in a.LastSessionID
    // 7. Return accumulated answer
}
```

### 2A.4 Add Factory Case

File: `agent/cli/provider/agent_provider.go`

```go
case registry.AgentRunner<Runner>:
    // Resolve path, construct Agent, return AgentRunner
```

---

## Phase 2B: TTY Runner Implementation

TTY runners follow a different architecture: they run inside a PTY managed by
`ttywatch`, and their output is captured via scrollback snapshots rather than
parsed from stdout JSON.

### 2B.1 Register the TTY Provider

File: `pkgs/agenttty/provider.go`

Add to `init()`:

```go
_ = Register(Provider{
    ID:                 "<runner>-tty",
    RegistryDir:        "<runner>-tty-registry",
    BannerProvider:     "<runner>",
    DisableTail:        true,  // unless you implement StartEventTail
    BuildArgv:          Build<Runner>CommandArgv,
    DetectScreenStatus: detect<Runner>ScreenStatus,
    CheckWritable:      check<Runner>Writable,
})
```

Fields:

| Field | Purpose |
|-------|---------|
| `ID` | Runner identifier, e.g. `"commandcode-tty"` |
| `RegistryDir` | Subdirectory under `AGENT_RUN_HOME` for session registry files |
| `BannerProvider` | Used by banner detection and scrollback cleanup |
| `DisableTail` | `true` if no structured sidecar stream (most mock-based runners) |
| `BuildArgv` | Builds the command + flags to run inside the PTY |
| `DetectScreenStatus` | Classifies scrollback as `"starting"`, `"idle"`, `"busy"` |
| `CheckWritable` | Reports whether the PTY accepts injected input |
| `StartEventTail` | Optional: tail structured events from runner output |

### 2B.2 Implement BuildArgv

Create `pkgs/agenttty/<runner>.go`:

```go
const env<Runner>TTYCommand = "AGENT_RUN_<RUNNER>_TTY_COMMAND"

func Build<Runner>CommandArgv(env *exec.Env, settingsPath,
    agentRunnerBinary, model, resumeSession string) ([]string, error) {
    // Allow env var override (for fake TUI in tests)
    if hook := strings.TrimSpace(os.Getenv(env<Runner>TTYCommand)); hook != "" {
        return parseShellWords(hook)
    }
    // Resolve binary from --agent-runner-binary
    spec := strings.TrimSpace(agentRunnerBinary)
    if spec == "" {
        return nil, fmt.Errorf("<runner>-tty: --agent-runner-binary is required")
    }
    words, err := parseShellWords(spec)
    if err != nil {
        return nil, err
    }
    return words, nil
}
```

**Headless vs open mode:** `BuildArgv` should return the command + flags, but
should NOT inject print-mode flags like `-p`. Whether `-p` is added is decided
in `run.go` based on `opts.Open`:

```go
// In run.go, after BuildArgv returns:
if runnerID == "commandcode-tty" && !opts.Open {
    argv = append(argv, "-p", p)  // headless: non-interactive, exits cleanly
} else {
    argv = append(argv, p)        // open: interactive, PTY survives for attach
}
```

### 2B.3 Implement Banner Detection

```go
func detect<Runner>ScreenStatus(scrollback []byte) string {
    plain := stripPlain(scrollback)
    if strings.TrimSpace(plain) == "" {
        return "starting"
    }
    return "idle"
}
```

For runners that need to wait for a specific banner, return `"starting"` until
the marker appears. The banner detection system polls scrollback every 75ms with
a 30s timeout.

**Add banner detection** in `pkgs/agenttty/banner.go`:

```go
if provider == "<runner>" {
    return strings.TrimSpace(plain) != ""  // any content = ready
}
```

### 2B.4 Implement Writable Check

```go
func check<Runner>Writable(scrollback []byte) WritableStatus {
    return WritableStatus{Ready: true}
}
```

Return `Ready: true` when the PTY accepts input (most CLI runners are always
writable once the banner appears).

### 2B.5 Handle Turn Completion

TTY runners use `waitForPersistentTurnRemote` (keep-alive) or `WaitHeadless`
(one-shot). Turn completion checks if the extracted scrollback text is non-empty:

```go
// In banner.go, persistentTurnComplete():
if cfg.runnerID == "<runner>-tty" {
    captured := extractAssistantTextForProvider(scrollback, prompt,
        cfg.bannerMarkers, cfg.bannerProvider)
    return strings.TrimSpace(captured) != ""
}
```

**Keep-alive policy:** TTY runners should always set `KeepTerminalAlive: true`
so the PTY survives after the agent exits — enabling `snapshot`, `attach`, and
`send` on the session:

```go
// In run.go:
if runnerID == "<runner>-tty" {
    opts.KeepTerminalAlive = true
}
```

### 2B.6 Clean Scrollback Output

TTY runners' output is captured from raw PTY scrollback. You must filter out
TUI chrome (logo, separators, status lines). Add a provider-specific cleaner
in `pkgs/agenttty/capture.go`:

```go
func extractAssistantTextForProvider(scrollback []byte, prompt string,
    markers []string, provider string) string {
    // ... existing checks for SUBMITTED:/Response: ...

    if provider == "<runner>" {
        return clean<Runner>Scrollback(scrollback, prompt)
    }
    // ... fallback generic cleaner ...
}

func clean<Runner>Scrollback(scrollback []byte, prompt string) string {
    plain := strings.TrimSpace(stripPlain(scrollback))
    lines := strings.Split(plain, "\n")
    var kept []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        if skip<Runner>TuiLine(line, prompt) {
            continue
        }
        kept = append(kept, line)
    }
    return strings.TrimSpace(strings.Join(kept, "\n"))
}

func skip<Runner>TuiLine(line, prompt string) bool {
    // Filter: logo art, separator bars, status lines, prompt echo,
    // permission bypass text, shortcut hints, retry/connection messages
    return false
}
```

### 2B.7 Emit Scrollback as Events

For runners without structured sidecar streams (`DisableTail: true`), add them
to the scrollback fallback emitter in `pkgs/agenttty/run.go`:

```go
if !streamed && opts.Emit != nil &&
    (runnerID == "codex-tty" || runnerID == "<runner>-tty") {
    text := strings.TrimSpace(captured)
    if text != "" {
        opts.Emit(types.AgentEvent{
            Type: types.ActionMessage,
            Role: "assistant",
            Text: text,
        })
    }
}
```

---

## Phase 3: llm-mock Integration (for Command Code-like runners)

If your runner needs a mock HTTP backend for testing, follow the
`llm-mock-run-commandcode` pattern.

### 3.1 Create the Orchestrator

File: `agent/llm/llm-mock/run/<runner>.go`

```go
func Run<Runner>(args []string, opts Run<Runner>Options) error {
    // 1. Load config
    loaded, err := mockconfig.LoadMerged("")
    // 2. Reserve port, start in-process mock server
    srv, err := server.Start(ctx, server.Options{Config: loaded, ...})
    // 3. Set sandbox env vars: HOME, API_URL, API_KEY
    // 4. Build and run the runner binary
    // 5. Wait, propagate exit code
    // 6. Cleanup
}
```

### 3.2 Create the Shortcut Binary

File: `agent/llm/llm-mock/llm-mock-run-<runner>/main.go`

A thin wrapper that calls `Run<Runner>()` with args from `os.Args[1:]`.

### 3.3 Implement Mock Endpoints

In `agent/llm/llm-mock/server/server.go`, add handlers for the runner's API
endpoints:

```go
registerHandler(mux, "/<runner>/generate", handler.handleRunnerGenerate, httpLog)
registerHandler(mux, "/<runner>/whoami", handler.handleRunnerWhoami, httpLog)
```

**Common pitfall:** API request bodies may differ between headless/non-interactive
mode and interactive mode. Use `json.RawMessage` for fields that might be string
or array:

```go
type runnerMessage struct {
    Role    string          `json:"role"`
    Content json.RawMessage `json:"content"`  // string OR array of blocks
}
```

---

## Phase 4: Testing

### 4.1 Doctest Tree

Create `cmd/agent-run/tests/<runner>-tty/` (for TTY runners) or
`agent/cli/<runner>/tests/` (for CLI runners).

Minimal test structure:

```
cmd/agent-run/tests/<runner>-tty/
├── DOCTEST.md          # DSN, version 0.0.2, test index, Request/Response/Run
├── SETUP.md            # Build agent-run + mock binary, set AGENT_RUN_HOME
├── run-headless/       # Basic JSON output test
│   ├── SETUP.md
│   └── ASSERT.md
├── open-keep-alive/    # --open keeps PTY alive
│   ├── SETUP.md
│   └── ASSERT.md
└── send-followup/      # Send to live session
    ├── SETUP.md
    └── ASSERT.md
```

Key test patterns:

1. **SETUP.md** must set `req.RepoRoot` from `DOCTEST_ROOT`, set `req.TempDir =
   t.TempDir()`, and build with `build.Dir = req.RepoRoot`.

2. **Doctest test** cache uses `go.work`; test runs from a generated cache
   directory — build paths must be relative to `req.RepoRoot`, not the test dir.

3. **Send followup test** is multi-step: start open session, run `send`, wait,
   snapshot, verify both prompts.

### 4.2 E2E Inspect Script

Create `script/debug/cmd-mock-inspect/main.go` for quick manual verification
without the doctest harness. Uses `ttywatch.EphemeralSession` to run the mock
in a PTY, send input, and verify output.

### 4.3 Run

```sh
# Vet
doctest vet ./cmd/agent-run/tests/<runner>-tty

# Run all
doctest test ./cmd/agent-run/tests/<runner>-tty

# Run specific
doctest test -v ./cmd/agent-run/tests/<runner>-tty/run-headless

# Run inspect script
go run ./script/debug/cmd-mock-inspect /tmp/<runner>-mock
```

---

## Files Changed Checklist

### TTY Runner Path

| # | File | What |
|---|------|------|
| 1 | `pkgs/agenttty/provider.go` | Register provider in `init()` |
| 2 | `pkgs/agenttty/<runner>.go` | BuildArgv, banner/writable detection |
| 3 | `pkgs/agenttty/banner.go` | Banner detection case, turn completion case |
| 4 | `pkgs/agenttty/capture.go` | Scrollback cleaner for TUI chrome |
| 5 | `pkgs/agenttty/run.go` | Keep-alive policy, -p injection (headless), scrollback emit |
| 6 | `agent/cli/registry/settings_cli.go` | Runner ID constant |

### llm-mock Integration

| # | File | What |
|---|------|------|
| 7 | `agent/llm/llm-mock/run/<runner>.go` | Orchestrator (mock server + runner child) |
| 8 | `agent/llm/llm-mock/llm-mock-run-<runner>/main.go` | Shortcut binary |
| 9 | `agent/llm/llm-mock/server/server.go` | API endpoint handlers + content type fix |
| 10 | `agent/llm/llm-mock/main.go` | Same for standalone binary (duplicated types) |

### Testing

| # | File | What |
|---|------|------|
| 11 | `cmd/agent-run/tests/<runner>-tty/DOCTEST.md` | Test tree root |
| 12 | `cmd/agent-run/tests/<runner>-tty/SETUP.md` | Build + AGENT_RUN_HOME setup |
| 13 | `cmd/agent-run/tests/<runner>-tty/run-headless/` | Headless run test |
| 14 | `cmd/agent-run/tests/<runner>-tty/open-keep-alive/` | Open + snapshot test |
| 15 | `cmd/agent-run/tests/<runner>-tty/send-followup/` | Send E2E test |
| 16 | `script/debug/cmd-mock-inspect/main.go` | Standalone inspect tool |

---

## Common Pitfalls

1. **Build paths in doctest SETUP.md**: Tests run from a cache dir, not the repo
   root. Always use `build.Dir = req.RepoRoot` and resolve `req.RepoRoot` from
   `filepath.Join(DOCTEST_ROOT, "../../../..")`.

2. **Mock request format differences**: Interactive mode often sends different
   body shapes than headless mode (e.g., content as array of blocks vs string).
   Use `json.RawMessage` + `extractContentText()` helpers.

3. **Keep-alive vs one-shot**: TTY runners need `KeepAlive: true` for
   `HeadlessRun()` so the PTY survives for snapshot/attach/send.

4. **Scrollback capture timing**: If the runner exits fast, capture the snapshot
   before `WaitHeadless` returns. Pre-capture with `waitContentAppears` helper.

5. **TUI chrome in captured text**: Unfiltered PTY scrollback contains ANSI escape
   codes, logo art, separator bars, status messages, and keyboard shortcut hints.
   Write a provider-specific cleaner that strips all of these.

6. **`-p` injection for non-interactive mode**: Some runners need a `-p`/`--print`
   flag to exit after responding. Inject it only for headless mode (not `--open`).

7. **Send queue drainer**: `agent-run send` enqueues, starts a drainer that polls
   `WaitUntilWritable` then calls `SendMessage`. The drainer must run while the
   PTY session is alive — keep-alive is essential.
