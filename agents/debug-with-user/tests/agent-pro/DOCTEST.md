# agent-pro debug-with-user Registration Tests

Doc-style tests verifying `debug-with-user` is registered in `agent-pro skills`.
This is a **nested DOCTEST root** — it builds `agent-pro` and exercises the
skill show path (integration proxy for `knownSkills` registration).

# DSN (Domain Specific Notion)

**agent-pro** maintains a `knownSkills` map in `cmd/agent-pro/skill_cmd.go`.
Each skill name maps to embedded SKILL.md content. `agent-pro skill <name> show`
prints that content; unknown names error with the available list.

Registration of `debug-with-user` means `agent-pro skill debug-with-user --show`
succeeds and prints frontmatter `name: debug-with-user`.

## Version

0.0.2

## Decision Tree

```
agent-pro/
├── DOCTEST.md
├── SETUP.md                           # build agent-pro binary
└── register/                          # skill show succeeds for debug-with-user
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `register` | `agent-pro skill debug-with-user --show` → contains `name: debug-with-user` |

## How to Run

```sh
doctest vet ./agents/debug-with-user/tests/agent-pro
doctest test -v ./agents/debug-with-user/tests/agent-pro/register
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
	"testing"
	"time"
)

type Request struct {
	RepoRoot  string
	TempDir   string
	AgentPro  string
	Args      []string
	Env       []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.AgentPro, req.Args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)

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
		resp.ExitCode = 0
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

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}
```
