# agent-hub home default — Doctests

Binary doctests for how `agent-hub` resolves its home directory: default
`$HOME/.agent-hub` vs `AGENT_HUB_HOME` override. Uses `daemon status` JSON
`home` field; no daemon is started.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — harness invoking the built `agent-hub` binary.
- **agent-hub CLI** — `cmd/agent-hub`; `daemon status` prints JSON including
  resolved `home` and `running`.
- **Environment** — `HOME` and optional `AGENT_HUB_HOME` on the **child**
  process only (`cmd.Env`).
- **Harness** — builds `./cmd/agent-hub` once per leaf temp dir; captures
  stdout/stderr/exit.

**Behaviors**

- With `AGENT_HUB_HOME` unset and isolated `HOME`, `home` is
  `$HOME/.agent-hub` and `running` is false.
- With `AGENT_HUB_HOME` set to a custom path, `home` equals that path
  (env wins over default).

## Version

0.0.1

## Decision Tree

```
home-default/
├── default-uses-home/   AGENT_HUB_HOME unset → $HOME/.agent-hub
└── env-override/        AGENT_HUB_HOME set → custom path
```

## How to Run

```sh
doctest test ./cmd/agent-hub/tests/home-default/...   # discovery (skips e2e)
doctest test --label e2e ./cmd/agent-hub/tests/home-default/...
doctest test --label-all ./cmd/agent-hub/tests/home-default/...
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Home        string
	Command     string
	Args        []string
	Env         []string
	TempDir     string
	AgentHub    string
	RepoRoot    string
	UserHomeDir string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	if req.Command == "" {
		req.Command = req.AgentHub
	}
	return execCmd(t, req.Command, req.Args, req.TempDir, req.Env, "")
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, stdin string) (*Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}
```
