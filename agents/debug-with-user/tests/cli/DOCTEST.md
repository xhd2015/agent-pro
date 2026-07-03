# debug-with-user CLI Tests

Doc-style tests for the `debug-with-user` CLI binary: `ask` with dry-run env
vars (CI-safe, no GUI) and `skill show`. This is a **nested DOCTEST root** —
it builds its own binary and does not inherit the parent dialog `Run`.

# DSN (Domain Specific Notion)

The **CLI** wraps the dialog package for subprocess use (os-bar, agents).
`debug-with-user ask` accepts title, message, repeated `--option` flags,
`--affirm`, and `--cancel`. It always appends **Customize** internally.

With `DEBUG_WITH_USER_DRY_RUN=1`, the CLI never invokes `osascript`. Staged
env vars simulate user choices:

- `DEBUG_WITH_USER_DRY_RUN_BUTTON` — step-1 button label (including `Customize`)
- `DEBUG_WITH_USER_DRY_RUN_TEXT` — step-2 typed text when Customize is chosen
- `DEBUG_WITH_USER_DRY_RUN_DISMISSED=1` — user cancels (exit 1)

Success prints one JSON line on stdout. `skill show` prints embedded SKILL.md.

## Version

0.0.2

## Decision Tree

```
cli/
├── DOCTEST.md
├── SETUP.md                           # build debug-with-user binary
├── ask/
│   ├── dry-run/
│   │   ├── button-affirmed/
│   │   ├── button-not-affirmed/
│   │   ├── customize/
│   │   └── dismissed/
│   └── non-mac/
└── skill-show/
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `ask/dry-run/button-affirmed` | Affirm preset via dry-run → `via=button`, `affirmed=true`, exit 0 |
| 2 | `ask/dry-run/button-not-affirmed` | Non-affirm preset → `affirmed=false` |
| 3 | `ask/dry-run/customize` | Customize + text → `via=free_text`, no `affirmed` field |
| 4 | `ask/dry-run/dismissed` | Cancel/dismiss → exit 1 |
| 5 | `ask/non-mac` | No dry-run on non-darwin → exit 2, macOS-only error |
| 6 | `skill-show` | `skill show` output contains skill frontmatter name |

## How to Run

```sh
doctest vet ./agents/debug-with-user/tests/cli
doctest test -v ./agents/debug-with-user/tests/cli
doctest test -v ./agents/debug-with-user/tests/cli/ask/dry-run/button-affirmed
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Binary   string
	Args     []string
	Env      []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	JSON     map[string]any
	Err      error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Binary, req.Args...)
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
		parseStdoutJSON(t, resp)
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

func parseStdoutJSON(t *testing.T, resp *Response) {
	t.Helper()
	line := strings.TrimSpace(resp.Stdout)
	if line == "" || !strings.HasPrefix(line, "{") {
		return
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("invalid JSON stdout: %v\n%s", err, resp.Stdout)
	}
	resp.JSON = obj
}

func defaultAskArgs() []string {
	return []string{
		"ask",
		"--title", "Step 1 — Did VS Code open?",
		"--message", "Project folder:\n/tmp/demo",
		"--option", "Yes — window opened",
		"--option", "No — window did not open",
		"--affirm", "Yes — window opened",
		"--cancel", "Cancel",
	}
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func assertJSONField(t *testing.T, resp *Response, key string, want any) {
	t.Helper()
	if resp.JSON == nil {
		t.Fatalf("expected JSON on stdout, got:\n%s", resp.Stdout)
	}
	got, ok := resp.JSON[key]
	if !ok {
		t.Fatalf("JSON missing key %q: %v", key, resp.JSON)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("JSON[%q] = %v (%T), want %v", key, got, got, want)
	}
}

func assertJSONNoKey(t *testing.T, resp *Response, key string) {
	t.Helper()
	if resp.JSON == nil {
		return
	}
	if _, ok := resp.JSON[key]; ok {
		t.Fatalf("JSON should not contain %q: %v", key, resp.JSON)
	}
}

func assertBoolField(t *testing.T, resp *Response, key string, want bool) {
	t.Helper()
	if resp.JSON == nil {
		t.Fatalf("expected JSON on stdout, got:\n%s", resp.Stdout)
	}
	got, ok := resp.JSON[key]
	if !ok {
		t.Fatalf("JSON missing key %q: %v", key, resp.JSON)
	}
	b, ok := got.(bool)
	if !ok {
		t.Fatalf("JSON[%q] = %v (%T), want bool", key, got, got)
	}
	if b != want {
		t.Fatalf("JSON[%q] = %v, want %v", key, b, want)
	}
}
```
