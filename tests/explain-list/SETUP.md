# Scenario

**Feature**: explain list reads isolated session store and prints cards without LLM

```
# build explain once; isolate debug config home; seed sessions; run list
caller: explain list [--limit N] [--grep P]... [--or|--and] [--color]
  -> session store under $AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME/sessions
  -> optional body grep filter (OR default / AND) then limit; pretty cards
  -> no EXPLAIN_AGENT_PATH start
doctest <- stdout / stderr / exit code
```

## Preconditions

- `go` is available in PATH (to build `./cmd/explain`). Tests skip otherwise.
- `cmd/explain` exists; implementer adds `list` dispatch + formatting (RED until then).
- Session cache dir: `$TMPDIR/explain-list-doctest-<d.DOCTEST_SESSION_ID>/`
  (shared `explain` binary + fake agent stub across parallel leaves).
- Per-leaf isolation: each leaf gets its own `ConfigHome` under `t.TempDir()`.
- Parent env `NO_COLOR` / `FORCE_COLOR` / `CLICOLOR_FORCE` are stripped in `Run`
  so color leaves control the policy explicitly.
- Fake agent at `EXPLAIN_AGENT_PATH` exits 99 with `FAKE_AGENT_INVOKED` on stderr
  if ever started.

## Steps

1. Resolve module root from `d.DOCTEST_ROOT` upward (`go.mod` + `cmd/explain`).
2. Build `explain` once per session into the session cache (file-locked).
3. Ensure fake agent stub exists in the same cache.
4. Create isolated `ConfigHome` for this leaf; default `Args` to `["list"]`.
5. Leaves append `Sessions`, refine `Args` / `EnvExtra`; `Run` seeds and executes.

## Context

- Debug env: `AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME` (not HOME).
- Session time ordering uses dirname prefix `YYYY-MM-DD-HH-mm-ss`, not mtime.
- User-facing stdout must end with a trailing newline after the last content line.
- Corrupt session dirs: create with `AgentRunner == "" && Model == "" && Messages == nil`
  so harness makes the directory but writes no `session.data`.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("skipping: go not found in PATH")
	}
	repoRoot, err := findModuleRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot

	bin, err := buildExplainOnce(t, d)
	if err != nil {
		return err
	}
	req.Bin = bin

	fake, err := ensureFakeAgent(t, d)
	if err != nil {
		return err
	}
	req.FakeAgentPath = fake

	req.ConfigHome = filepath.Join(t.TempDir(), "explain-config-home")
	if err := os.MkdirAll(req.ConfigHome, 0o755); err != nil {
		return err
	}
	if len(req.Args) == 0 {
		req.Args = []string{"list"}
	}
	return nil
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func assertStdoutEndsWithNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with trailing newline; got %q", stdout)
	}
}

func assertNoANSI(t *testing.T, s string) {
	t.Helper()
	if strings.Contains(s, "\x1b") {
		t.Fatalf("expected no ANSI escapes, got:\n%s", s)
	}
}

func assertHasANSI(t *testing.T, s string) {
	t.Helper()
	if !strings.Contains(s, "\x1b") {
		t.Fatalf("expected ANSI escapes, got:\n%s", s)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("missing %q in:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("unexpected %q in:\n%s", substr, s)
	}
}

// bodyAfterLabel returns the message body after a "Q  " or "A  " label on a line,
// stripping a trailing newline. Available for leaves that inspect a single Q/A line.
func bodyAfterLabel(line, label string) string {
	line = strings.TrimRight(line, "\n")
	idx := strings.Index(line, label)
	if idx < 0 {
		return ""
	}
	return line[idx+len(label):]
}

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}
```
