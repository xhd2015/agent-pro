# Adding a New Agent Runner

This guide documents the full process of adding a new agent runner (e.g., crush, cursor)
to the agent-pro system, broken into 3 phases.

---

## Phase 1: Event Types (`agent/event/`)

### 1.1 Define Wire Format Types

Create `agent/event/<runner>_types/types.go` with the runner's event wire format.

For CLI-based runners, you typically need:
- **Event envelope**: top-level type + payload discriminator
- **Message parts**: text, reasoning, tool_call, tool_result, finish
- **Agent event**: errors, warnings, lifecycle signals
- **Run complete**: terminal event with session_id, text, error

Example (crush):
```go
// agent/event/crush_types/types.go
package crush_types

type Event struct {
    Type    EventType       `json:"type"`     // "message", "agent_event", "run_complete"
    Payload json.RawMessage `json:"payload"`
}

type MessagePayload struct {
    ID        string `json:"id"`
    Role      string `json:"role"`           // "assistant", "user", "system"
    SessionID string `json:"session_id"`
    Parts     []Part `json:"parts"`
}

type Part struct {
    Type PartType        `json:"type"`       // "text", "reasoning", "tool_call", etc.
    Data json.RawMessage `json:"data"`
}
```

### 1.2 Write Converters

Create `agent/event/<runner>_types/convert.go` with two functions:

- `To<Runner>(events []types.AgentEvent, sessionID string) []Event` — canonical → runner format
- `From<Runner>(events []Event, sessionID string) []types.AgentEvent` — runner format → canonical

These convert between the runner's wire format and the canonical `agent/event/types.AgentEvent`.

### 1.3 Add Doctests

Create `agent/event/tests/<runner>_types/` with a doctest tree testing:
- Each conversion direction (canonical → runner, runner → canonical)
- Each event type (thinking, text, tool call, error, done)
- Edge cases (non-assistant roles, mixed parts, empty payloads)

---

## Phase 2: CLI Agent (`agent/cli/`)

### 2.1 Register the Runner ID

File: `agent/cli/registry/settings_cli.go`

```go
// Add constant:
AgentRunner<Runner> AgentRunnerID = "<runner>"

// Add settings key:
<Runner>CLIPathSettingKey = "<runner>_cli_path"

// Add env var:
Env<Runner>CLIPath = "AGENT_RUNNER_<RUNNER>_PATH"

// Add field to Settings struct:
<Runner>CLIPath string `json:"<runner>_cli_path,omitempty"`

// Add case in LoadConfiguredStringSetting():
case <Runner>CLIPathSettingKey:
    return strings.TrimSpace(settings.<Runner>CLIPath)
```

### 2.2 Implement the Agent

Create `agent/cli/<runner>/<runner>.go` implementing `registry.Agent`:

```go
package <runner>

type <Runner>Agent struct {
    AgentPath    string
    SettingsPath string
    Workspace    string
    Env          *exec.Env
}

func FindAgentPath(env *exec.Env) (string, error) {
    // Look for "<runner>" binary in PATH
    if path, err := env.LookPath("<runner>"); err == nil {
        return path, nil
    }
    return "", fmt.Errorf("<runner> not found in PATH")
}

func (a *<Runner>Agent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
    // 1. Resolve binary path
    // 2. Build command args (flags, model, session, prompt)
    // 3. Spawn process
    // 4. Read stdout line by line
    // 5. Parse JSON events using <runner>_types
    // 6. Call onDelta for text, opts.OnToolCall for tool calls
    // 7. Write raw lines to opts.RawLog if set
    // 8. Return accumulated text
}

func (a *<Runner>Agent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
    // Return nil if the runner doesn't support model listing
    return nil, nil
}
```

Key patterns to follow (see `agent/cli/opencode/opencode.go`):
- Use `a.Env.CommandContext(ctx, ...)` to spawn the binary
- Use `bufio.Scanner` with large buffer for stdout
- Write raw lines to `opts.RawLog` for session logging
- Handle stream errors (authentication, rate limits, crashes)
- Support `opts.SessionID` for session resume
- Support `opts.DisableSubAgents` by appending to prompt
- Use `registry.ResolveConfiguredCLIPath()` for path resolution fallback chain

### 2.3 Add Factory Case

File: `agent/cli/provider/agent_provider.go`

```go
import <runner>agent "github.com/xhd2015/agent-pro/agent/cli/<runner>"

// In Build():
case registry.AgentRunner<Runner>:
    <runner>Path, err := registry.ResolveConfiguredCLIPath(
        settingsPath,
        registry.<Runner>CLIPathSettingKey,
        registry.Env<Runner>CLIPath,
        "",
        func() (string, error) { return <runner>agent.FindAgentPath(env) },
    )
    if err != nil {
        return registry.AgentRunner{}, fmt.Errorf("<runner> not found: %w", err)
    }
    return registry.AgentRunner{
        ID:   registry.AgentRunner<Runner>,
        Name: "<Runner Display Name>",
        Agent: &<runner>agent.<Runner>Agent{
            AgentPath:    <runner>Path,
            SettingsPath: settingsPath,
            Workspace:    workspace,
            Env:          env,
        },
    }, nil
```

### 2.4 Add Doctests

Create `agent/cli/<runner>/tests/` with integration tests covering:
- Basic query (fresh session, verify answer)
- Session resume (multi-turn, verify context retention)
- Session ID capture (verify `LastSessionID` is populated)
- Binary not found (skip/error)

The `Response` struct must include `SessionID string` populated from `agent.LastSessionID`.

---

## Phase 3: Integrate with `agent/agentui/runner/`

### 3.1 Runner Dispatch

File: `agent/agentui/runner/runner.go`

The `RunLLM()` and `RunPlain()` functions check `agentRunner != "opencode"` to route non-opencode runners through `askAgentCLI()`. If your runner is non-opencode, it's automatically routed — but verify:

1. `askAgentCLI()` calls `agentprovider.Build()` which resolves your factory case
2. `sessionLogWriter` captures session IDs in your runner's format (check `agent/subagent/log_writer.go`)
3. `HandleYieldPendingQuestions()` and `HandleReportProgress()` work correctly

If your runner needs special handling (different from the generic `askAgentCLI()` path), add a case:

```go
case "<runner>":
    output, err := handle<Runner>CLI(prompt, llmModel, sessionID, sessionDir)
    doneCh <- Done{Output: output, Err: err}
```

### 3.2 CLI Validation

File: `agent/agentui/run_options.go`

Add `<runner>` to the accepted `--agent-runner` values:

```go
if agentRunner != "opencode" && agentRunner != "codex" && agentRunner != "fake-codex" && agentRunner != "<runner>" {
    return runOptions{}, fmt.Errorf("unsupported agent runner: %s (supported: opencode, codex, fake-codex, <runner>)", agentRunner)
}
```

### 3.3 Trace Adapter

Create `agent_trace/<runner>/<runner>.go` registering a trace parser via `init()`:

```go
package <runner>

import (
    "encoding/json"
    "github.com/xhd2015/agent-pro/agent_trace/types"
)

type <runner>TraceAdapter struct{}

func init() {
    // Priority: pick a unique value. Convention:
    //   codex=10, crush=12, opencode=15, cursor=20
    types.RegisterAgentTraceAdapter(20, <runner>TraceAdapter{})
}

func (<runner>TraceAdapter) Name() string { return "<runner>" }

func (<runner>TraceAdapter) Parse(raw json.RawMessage) (types.AgentTraceParsedEvent, bool) {
    // Parse the raw event line using <runner>_types
    // Return a normalized AgentTraceParsedEvent with Message and/or Activity
}
```

Then add a blank import in `agent_trace/agent_trace.go`:

```go
_ "github.com/xhd2015/agent-pro/agent_trace/<runner>"
```

### 3.4 Agent Hub Hooks

File: `cmd/agent-hub/main.go`

Add runner to accepted values and map event names:

```go
// In normalizeHook():
case "codex", "fake-codex", "opencode", "fake-opencode", "<runner>":

// In hookEventType():
case "<runner>":
    switch nativeEvent {
    case "session.created":
        return model.EventSessionStarted, true
    case "message.created":
        return model.EventPromptSubmitted, true
    // ... etc
    }
```

### 3.5 Session/Subagent Integration (Optional)

Files under `agent/subagent/`:

- `subagent.go:116` — Change default runner from `"opencode"` to `<runner>`
- `log_writer.go` — Add `<runner>` format session ID extraction if different from opencode
- `session.go` — Add `<runner>_session_id` field in meta, update `readOpencodeSessionID` equivalent

---

## Checklist

| # | Phase | Item | File |
|---|-------|------|------|
| 1 | 1 | Wire format types | `agent/event/<runner>_types/types.go` |
| 2 | 1 | Converter functions | `agent/event/<runner>_types/convert.go` |
| 3 | 1 | Doctests | `agent/event/tests/<runner>_types/` |
| 4 | 2 | Runner ID constant | `agent/cli/registry/settings_cli.go` |
| 5 | 2 | Settings keys & env vars | `agent/cli/registry/settings_cli.go` |
| 6 | 2 | Agent implementation | `agent/cli/<runner>/<runner>.go` |
| 7 | 2 | Factory case | `agent/cli/provider/agent_provider.go` |
| 8 | 2 | Doctests | `agent/cli/<runner>/tests/` |
| 9 | 3 | Runner dispatch | `agent/agentui/runner/runner.go` |
| 10 | 3 | CLI validation | `agent/agentui/run_options.go` |
| 11 | 3 | Trace adapter | `agent_trace/<runner>/<runner>.go` |
| 12 | 3 | Trace import | `agent_trace/agent_trace.go` |
| 13 | 3 | Agent hub hooks | `cmd/agent-hub/main.go` |
| 14 | 3 | Subagent defaults (opt) | `agent/subagent/` |
