# agent-pro verify-on-behalf-of-user Registration Tests

Doc-style tests verifying `verify-on-behalf-of-user` is registered in
`agent-pro skills`, exposes embedded SKILL.md and nested topics (including
scenario), and installs utility scripts.

# DSN (Domain Specific Notion)

**agent-pro** `knownSkills` includes `verify-on-behalf-of-user` with embedded
SKILL.md and `ExtraFiles` (scripts, templates, nested topics). `agent-pro skill
--install` writes SKILL.md plus bundled files under
`.agents/skills/verify-on-behalf-of-user/`.

## Version

0.0.2

## Decision Tree

```
agent-pro/
├── DOCTEST.md
├── SETUP.md
├── register/
├── list/
├── install/
└── scenario-topic/
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `register` | `--show` transcript topic + format rules; skill is registered |
| 2 | `list` | `agent-pro skills` lists skill with description |
| 3 | `install` | `--install` writes scripts, templates, scenario + transcript topics |
| 4 | `scenario-topic` | `--show` scenario topic: depth labels, browser-agent, FAIL rules |

## How to Run

```sh
doctest vet ./agents/verify-on-behalf-of-user/tests/agent-pro
doctest test -v ./agents/verify-on-behalf-of-user/tests/agent-pro/...
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
		t.Fatalf("exit code: got %d want %d\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}
```
