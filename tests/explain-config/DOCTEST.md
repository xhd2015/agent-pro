# explain config preferences (CLI)

E2E doctests for `explain --set-config`, `--show-config`, and `--no-config`.
Persisted preferences live at `$AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME/config.json`.
Runner application without a working fake is asserted via `-v` stderr; override /
`--no-config` happy paths use the working fake-opencode script.

# DSN (Domain Specific Notion)

```
explain --set-config --agent-runner R [--model M]
explain --show-config
explain [--no-config] [--agent-runner R] <msg>
  precedence: CLI flag > config.json > opencode
```

## Version

0.0.1

## Decision Tree

```
explain-config/
├── DOCTEST.md
├── SETUP.md
├── help/                      # -h documents config flags
├── show/
│   └── empty/                 # missing file → {}
├── set/
│   ├── agent-runner/          # writes agent_runner
│   ├── requires-pref/         # bare --set-config → error
│   └── mutex-with-show/       # --set-config --show-config → error
├── apply/
│   ├── from-config/           # config=codex → notice + "codex" in stderr
│   ├── flag-overrides/        # CLI runner wins; no notice
│   └── no-config/             # --no-config → opencode; no notice
├── notice/
│   ├── color-force/           # --color → gray notice: prefix
│   └── conflict/              # --color + --no-color → error
└── corrupt/
    └── show-fails/            # bad JSON → non-zero Error:
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help` | help mentions set/show/no-config |
| 2 | `show/empty` | missing config → `{}` |
| 3 | `set/agent-runner` | persist `agent_runner=codex` |
| 4 | `set/requires-pref` | bare set-config errors |
| 5 | `set/mutex-with-show` | set+show mutually exclusive |
| 6 | `apply/from-config` | config runner + notice line |
| 7 | `apply/flag-overrides` | CLI runner wins; no notice |
| 8 | `apply/no-config` | skip file → built-in opencode; no notice |
| 9 | `notice/color-force` | `--color` grays `notice:` |
| 10 | `notice/conflict` | `--color` + `--no-color` errors |
| 11 | `corrupt/show-fails` | parse error on --show-config |

## How to Run

```sh
doctest vet ./tests/explain-config
doctest test -v ./tests/explain-config/...
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
	WorkingAgentPath string
	RepoRoot         string
	EnvExtra         []string
}

type Response struct {
	ExitCode     int
	Stdout       string
	Stderr       string
	ConfigJSON   string
	AgentRunner  string
	SessionMeta  json.RawMessage
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

	if raw, readErr := os.ReadFile(filepath.Join(req.ConfigHome, "config.json")); readErr == nil {
		resp.ConfigJSON = string(raw)
	}

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
