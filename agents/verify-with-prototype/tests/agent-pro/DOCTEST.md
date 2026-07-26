# agent-pro verify-with-prototype Registration Tests

Doc-style tests verifying `verify-with-prototype` is registered in `agent-pro skills`.
This is a **nested DOCTEST root** — it builds `agent-pro` and exercises skill show/list
paths (integration proxy for `knownSkills` registration).

# DSN (Domain Specific Notion)

**agent-pro** maintains a `knownSkills` map in `cmd/agent-pro/skill_cmd.go`.
Each skill name maps to embedded SKILL.md content. `agent-pro skill <name> show`
prints that content; unknown names error with the available list.
`agent-pro skills` (no args) lists every known skill with its frontmatter
description.

Registration of `verify-with-prototype` means `agent-pro skill verify-with-prototype --show`
succeeds and prints frontmatter `name: verify-with-prototype`; the skills list includes
`verify-with-prototype` with description text.

## Version

0.0.1

## Decision Tree

```
agent-pro/
├── DOCTEST.md
├── SETUP.md                           # build agent-pro binary
├── register/                          # skill show succeeds for verify-with-prototype
└── list/                              # agent-pro skills lists verify-with-prototype
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `register` | `agent-pro skill verify-with-prototype --show` → contains `name: verify-with-prototype` |
| 2 | `list` | `agent-pro skills` → lists `verify-with-prototype` with description |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./agents/verify-with-prototype/tests/agent-pro/...                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./agents/verify-with-prototype/tests/agent-pro/...
doctest test --label-all ./agents/verify-with-prototype/tests/agent-pro/...

doctest vet ./agents/verify-with-prototype/tests/agent-pro
doctest test -v ./agents/verify-with-prototype/tests/agent-pro/...
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

	"github.com/xhd2015/doctest/session"
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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
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