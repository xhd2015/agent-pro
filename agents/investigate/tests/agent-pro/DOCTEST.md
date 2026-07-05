# agent-pro investigate Registration Tests

Doc-style tests verifying `investigate` is registered in `agent-pro skills` and
referenced by `intent-route`. This is a **nested DOCTEST root** — it builds
`agent-pro` and exercises skill show/list paths (integration proxy for
`knownSkills` registration and intent-route wiring).

# DSN (Domain Specific Notion)

**agent-pro** maintains a `knownSkills` map in `cmd/agent-pro/skill_cmd.go`.
Each skill name maps to embedded SKILL.md content. `agent-pro skill <name> show`
prints that content; unknown names error with the available list.
`agent-pro skills` (no args) lists every known skill with its frontmatter
description.

**intent-route** is another embedded skill whose SKILL.md classifies user input
into categories and names a guideline command per category. After adding
**Investigation**, its show output must reference `investigate` and
`agent-pro skill investigate show`.

Registration of `investigate` means `agent-pro skill investigate show`
succeeds and prints frontmatter `name: investigate`; the skills list includes
`investigate` with description text; intent-route show mentions the
Investigation category and the investigate guideline.

## Version

0.0.2

## Decision Tree

```
agent-pro/
├── DOCTEST.md
├── SETUP.md                           # build agent-pro binary
├── register/                          # skill show succeeds for investigate
├── list/                              # agent-pro skills lists investigate
└── intent-route/                      # intent-route show references investigate
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `register` | `agent-pro skill investigate show` → contains `name: investigate` |
| 2 | `list` | `agent-pro skills` → lists `investigate` with description |
| 3 | `intent-route` | `agent-pro skill intent-route show` → Investigation category + investigate guideline |

## How to Run

```sh
doctest vet ./agents/investigate/tests/agent-pro
doctest test -v ./agents/investigate/tests/agent-pro/...
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