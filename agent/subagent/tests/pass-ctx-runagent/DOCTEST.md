# Pass Context Through runAgent

Verify that `runAgent` accepts `ctx context.Context` as its first parameter (replacing the hardcoded 30-second timeout), and that the context from the caller propagates to `runner.Agent.Ask`.

## Decision Tree

```
pass-ctx-runagent/
├── DOCTEST.md                # This file
├── SETUP.md                  # Root: Package under test = agent/subagent
│                              #   Run() calls runAgent(ctx, ...) directly
│
└── ctx-propagation/          # Context propagation behavior
    ├── SETUP.md              # Sets AgentRunner = "fake-codex"
    ├── canceled/             # Pre-canceled context → error
    │   ├── SETUP.md          # CancelCtx = true
    │   └── ASSERT.md         # err != nil
    └── not-canceled/         # Normal context → output
        ├── SETUP.md          # Prompt = "write hello world"
        └── ASSERT.md         # err == nil && output != ""
```

Branches:
- **Root**: Split on context state (canceled vs normal).
- **Group**: `ctx-propagation` narrows to cancellation behavior.
- **Leaves**: One for each context state.

## Test Index

| Leaf | Description |
|------|-------------|
| `ctx-propagation/canceled` | Pre-canceled context passed to `runAgent` → returns error |
| `ctx-propagation/not-canceled` | Normal `context.Background()` passed to `runAgent` → returns output |

## How to Run

```sh
# Validate tree structure
doctest vet ./external/agent-pro/agent/subagent/tests/pass-ctx-runagent/

# Run tests (expect RED — compilation fails until runAgent signature is updated)
doctest test -v ./external/agent-pro/agent/subagent/tests/pass-ctx-runagent/
```

```go
import (

    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
    "github.com/xhd2015/doctest/session"
)


type Request struct {
    AgentRunner string
    Prompt      string
    CancelCtx   bool
    SessionID   string
}

type Response struct {
    Output string
    Err    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
    // Residual: agentprovider resolves fake-codex via os.Getenv(AGENT_RUNNER_FAKE_CODEX_PATH)
    // / PATH LookPath with no Config/Options path on runAgent. Process Setenv + restore
    // until product accepts an explicit AgentPath on the runAgent test export.
    fakeCodexPath := filepath.Join(t.TempDir(), "fake-codex")
    moduleRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", "..", ".."))
    build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", fakeCodexPath, "./fake-codex")
    build.Dir = moduleRoot
    if out, err := build.CombinedOutput(); err != nil {
        return nil, fmt.Errorf("build fake-codex: %w\n%s", err, out)
    }
    prev, had := os.LookupEnv("AGENT_RUNNER_FAKE_CODEX_PATH")
    if err := os.Setenv("AGENT_RUNNER_FAKE_CODEX_PATH", fakeCodexPath); err != nil {
        return nil, err
    }
    defer func() {
        if had {
            _ = os.Setenv("AGENT_RUNNER_FAKE_CODEX_PATH", prev)
        } else {
            _ = os.Unsetenv("AGENT_RUNNER_FAKE_CODEX_PATH")
        }
    }()

    prompt := req.Prompt
    if prompt == "" {
        prompt = "write hello world"
    }

    ctx := context.Background()
    if req.CancelCtx {
        var cancel context.CancelFunc
        ctx, cancel = context.WithCancel(ctx)
        cancel()
    }

    rawLog := subagent.TestExported_NewSessionLogWriter()
    output, err := subagent.TestExported_runAgent(ctx, req.AgentRunner, "", prompt, req.SessionID, rawLog)
	return &Response{Output: output, Err: err}, err
}
```
