# slack-msg agent wire → agentrunapi (interactive open)

Classic TDD source-wire doctests for plan phase **P4**: migrate
`cmd/slack-msg` thread interactive open from `agentrunbridge.RunInteractiveOpen`
to **`agentrunapi.AutoSendOrResume`** (+ OpenInNewTerminal / WaitReady as needed).
Empty driver / empty `SLACK_LISTEN_AGENT_RUN` falls back to **agent-run**.

Does **not** rewrite the large `tests/slack-msg/` suite (regression after implement).

**Out of scope:** full slack-msg rewrite, local-bot, deleting agentrunbridge,
new slack-msg helper binary.

# DSN (Domain Specific Notion)

**Participants**

- **`cmd/slack-msg/agent.go`** — production glue between Slack listen/session
  routing and agent-run. Two entrypoints:
  - **`runAgentInteractiveOpen`** — thread-mode interactive open (TTY owns session).
  - **`runAgentStateless`** — one-shot with stdout capture for channel PostMessage.
- **`agentrunapi`** — in-process library: `AutoSendOrResume`, optional
  `OpenInNewTerminal` / `BuildFollowUpCommand` / `WaitReady`. Interactive open
  path must use this instead of bridge open profile.
- **`agentrunbridge`** — may **remain** for **`runAgentStateless`** only
  (`Run` + `Stateless` + `CaptureStdout`), until agentrunapi grows stateless
  capture. Must **not** be used for interactive open (`RunInteractiveOpen` gone).
- **`SLACK_LISTEN_AGENT_RUN` (`envAgentRun`)** — optional absolute/path binary.
  Empty → library FollowUp / driver defaults to **`agent-run`** (compat).
  Non-empty → that value is `DriverBinary` / binary for launch paths.
- **Existing `tests/slack-msg/listen/**`** — integration regression; stay GREEN
  after wire (not part of this tree).

**Behaviors**

```
# interactive open (thread mode) — P4 target
runAgentInteractiveOpen(prompt, sessionID, opts)
  -> agentrunapi.AutoSendOrResume(Opts{
       SessionID, Prompt, WorkspaceDir, AgentRunner, Env, …
       Open:true, NewTerminal:true (or OpenInNewTerminal path),
       AllowRelocateResumeSessionDir:true,
       DriverBinary: agentRunBinary()  // empty → library "agent-run"
     })
  -> optional WaitReady after launch
  -> NOT agentrunbridge.RunInteractiveOpen

# driver binary
agentRunBinary() = getenv("SLACK_LISTEN_AGENT_RUN")
  empty  -> DriverBinary "" -> agentrunapi FollowUp defaults "agent-run"
  set    -> DriverBinary = env value

# stateless (out of interactive cutover; keep bridge OK)
runAgentStateless(prompt, opts)
  -> agentrunbridge.Run(RunOpts{Stateless:true, CaptureStdout:true, …})
```

## Version

0.0.2

## Decision Tree

```
tests/slack-msg-agent-wire/
├── DOCTEST.md
├── SETUP.md
├── interactive/                         # runAgentInteractiveOpen cutover
│   ├── SETUP.md
│   ├── uses-agentrunapi/                # import agentrunapi + AutoSendOrResume
│   └── no-run-interactive-open/         # no RunInteractiveOpen (bridge open gone)
├── driver/                              # SLACK_LISTEN_AGENT_RUN / DriverBinary
│   ├── SETUP.md
│   ├── empty-defaults-agent-run/        # empty env → agent-run fallback present
│   └── env-sets-binary/                 # env const + agentRunBinary reads it
└── stateless/                           # keep bridge capture path
    ├── SETUP.md
    └── keeps-agentrunbridge-run/        # Stateless + CaptureStdout via bridge.Run
```

Parameter ranking (most → least significant):

1. **Entrypoint concern** — interactive open vs driver policy vs stateless
2. **Wire shape** — uses agentrunapi vs forbids bridge open
3. **Driver identity** — empty default agent-run vs env override

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `interactive/uses-agentrunapi` | `agent.go` imports `pkgs/agentrunapi` and references `AutoSendOrResume` |
| 2 | `interactive/no-run-interactive-open` | `agent.go` does not call/reference `RunInteractiveOpen` |
| 3 | `driver/empty-defaults-agent-run` | Empty-driver / empty-env fallback documents or defaults to `agent-run` |
| 4 | `driver/env-sets-binary` | `SLACK_LISTEN_AGENT_RUN` constant + `agentRunBinary` / getenv wiring |
| 5 | `stateless/keeps-agentrunbridge-run` | Stateless path still uses `agentrunbridge.Run` with Stateless/CaptureStdout |

## How to Run

```sh
# from agent-pro module root
doctest vet ./tests/slack-msg-agent-wire
doctest test ./tests/slack-msg-agent-wire

doctest test -v ./tests/slack-msg-agent-wire/interactive/uses-agentrunapi
doctest test -v ./tests/slack-msg-agent-wire/interactive/no-run-interactive-open
doctest test -v ./tests/slack-msg-agent-wire/driver/empty-defaults-agent-run
doctest test -v ./tests/slack-msg-agent-wire/driver/env-sets-binary
doctest test -v ./tests/slack-msg-agent-wire/stateless/keeps-agentrunbridge-run

# regression after implement (existing suite; not this tree)
doctest test ./tests/slack-msg/listen
```

Expect **RED** until `cmd/slack-msg/agent.go` interactive open uses agentrunapi.
Stateless may already pass if bridge Run remains. Empty-driver leaf RED until
fallback to `agent-run` is explicit in source (comment and/or `DriverBinary` /
default string).

### Target production shape (implementer; not implemented by designer)

```go
// cmd/slack-msg/agent.go (illustrative)

import (
	"os"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentrunbridge" // stateless only
)

const envAgentRun = "SLACK_LISTEN_AGENT_RUN"

func agentRunBinary() string {
	// Empty → agentrunapi DriverBinary "" defaults FollowUp to "agent-run".
	return os.Getenv(envAgentRun)
}

func runAgentInteractiveOpen(prompt, sessionID string, opts agentOptions) error {
	// Build store / opts; AutoSendOrResume with Open+NewTerminal (or OpenInNewTerminal),
	// DriverBinary: agentRunBinary(), AllowRelocateResumeSessionDir: true, Env, …
	return agentrunapi.AutoSendOrResume(ctx, agentrunapi.Opts{ /* … */ })
}

func runAgentStateless(prompt string, opts agentOptions) (string, error) {
	result, err := agentrunbridge.Run(agentrunbridge.RunOpts{
		Prompt: prompt, Binary: agentRunBinary(), Stateless: true, CaptureStdout: true,
		// …
	})
	return result.Stdout, err
}
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Request drives one leaf via Mode.
type Request struct {
	// Mode: interactive_api | interactive_no_bridge_open | driver_empty | driver_env | stateless_bridge
	Mode string
}

// Response holds source-scan observations for agent.go.
type Response struct {
	AgentSrc     string
	AgentPath    string
	ErrString    string
	// Convenience flags filled by Run (optional; Assert may re-scan AgentSrc).
	HasAgentrunapi       bool
	HasAutoSendOrResume  bool
	HasRunInteractiveOpen bool
	HasAgentrunbridge    bool
	HasBridgeRun         bool
	HasStateless         bool
	HasCaptureStdout     bool
	HasEnvAgentRun       bool
	HasAgentRunBinary    bool
	HasAgentRunDefault   bool // literal "agent-run" fallback / docs in source
	HasDriverBinary      bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	// DOCTEST_ROOT = tests/slack-msg-agent-wire → ../../cmd/slack-msg/agent.go
	path := filepath.Join(DOCTEST_ROOT, "..", "..", "cmd", "slack-msg", "agent.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read agent.go: %w", err)
	}
	src := string(data)
	resp.AgentSrc = src
	resp.AgentPath = path

	resp.HasAgentrunapi = strings.Contains(src, "pkgs/agentrunapi") || strings.Contains(src, "agentrunapi.")
	resp.HasAutoSendOrResume = strings.Contains(src, "AutoSendOrResume")
	resp.HasRunInteractiveOpen = strings.Contains(src, "RunInteractiveOpen")
	resp.HasAgentrunbridge = strings.Contains(src, "pkgs/agentrunbridge") || strings.Contains(src, "agentrunbridge.")
	resp.HasBridgeRun = strings.Contains(src, "agentrunbridge.Run(") || strings.Contains(src, "agentrunbridge.Run ")
	// Also accept RunOpts assignment style: agentrunbridge.Run(agentrunbridge.RunOpts{
	if strings.Contains(src, "agentrunbridge.Run") {
		resp.HasBridgeRun = true
	}
	resp.HasStateless = strings.Contains(src, "Stateless")
	resp.HasCaptureStdout = strings.Contains(src, "CaptureStdout")
	resp.HasEnvAgentRun = strings.Contains(src, "SLACK_LISTEN_AGENT_RUN")
	resp.HasAgentRunBinary = strings.Contains(src, "agentRunBinary")
	resp.HasAgentRunDefault = strings.Contains(src, `"agent-run"`) || strings.Contains(src, "agent-run")
	resp.HasDriverBinary = strings.Contains(src, "DriverBinary")

	switch req.Mode {
	case "interactive_api", "interactive_no_bridge_open", "driver_empty", "driver_env", "stateless_bridge":
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}
```
