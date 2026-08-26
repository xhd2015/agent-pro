# explain --agent-runner (CLI)

E2E doctests for `explain` agent-runner configuration via the built `explain`
binary. Help and validation need no LLM. One `opencode` happy-path leaf uses a
working fake agent via `EXPLAIN_AGENT_PATH`. Persistence for `codex` / `grok` /
`commandcode` is covered by L1 package tests with an injected mock runner.

# DSN (Domain Specific Notion)

**explain** accepts `--agent-runner` (`opencode` default, also `codex`, `grok`,
`commandcode`). Unknown values hard-error on stderr. Successful starts persist
`agent_runner` and `agent_runners_meta` under
`$AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME/sessions/`.

```
explain [--agent-runner R] <msg>
  -> validate R / help
  -> (optional) fake agent via EXPLAIN_AGENT_PATH
doctest <- exit / stdout / stderr / session.data
```

## Version

0.0.1

## Decision Tree

```
explain-agent-runner/
├── DOCTEST.md
├── SETUP.md
├── help/                    # -h lists supported runners
├── reject-unknown/          # unknown runner → non-zero + Error:
└── select/
    └── opencode/            # default runner + fake agent → session persisted
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help` | `-h` mentions opencode, codex, grok, commandcode |
| 2 | `reject-unknown` | `--agent-runner nope` → exit ≠0, unsupported error |
| 3 | `select/opencode` | prompt with fake opencode → `agent_runner=opencode` |

## How to Run

```sh
doctest vet ./tests/explain-agent-runner
doctest test -v ./tests/explain-agent-runner/...
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Bin              string
	Args             []string
	ConfigHome       string
	FakeAgentPath    string
	WorkingAgentPath string // optional EXPLAIN_AGENT_PATH override (working fake)
	RepoRoot         string
	EnvExtra         []string
}

type Response struct {
	ExitCode    int
	Stdout      string
	Stderr      string
	AgentRunner string
	SessionMeta json.RawMessage
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Bin == "" {
		t.Fatal("req.Bin not set")
	}
	if req.ConfigHome == "" {
		t.Fatal("req.ConfigHome not set")
	}
	if len(req.Args) == 0 {
		t.Fatal("req.Args empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	cmd.Dir = req.ConfigHome
	cmd.Env = buildEnv(req)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = ee.ExitCode()
		} else {
			return resp, err
		}
	}

	// Best-effort read of the newest session (if any).
	base := filepath.Join(req.ConfigHome, "sessions")
	entries, readErr := os.ReadDir(base)
	if readErr == nil && len(entries) > 0 {
		data, _ := readSessionFile(filepath.Join(base, entries[0].Name(), "session.data"))
		resp.AgentRunner = data.AgentRunner
		if data.AgentRunnersMeta != nil {
			resp.SessionMeta = data.AgentRunnersMeta[data.AgentRunner]
		}
	}
	return resp, nil
}
```
