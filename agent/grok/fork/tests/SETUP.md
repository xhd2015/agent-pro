# Scenario

**Feature**: grok-fork library Main with injectable ancestor resolve and launch

```
# harness fills Request (args, pid, procs, lsof, grok home, writers)
fork.Main(args, Options{…inject…})
  -> Mode A: ancestor + session + OpenInNewTerminal
  -> Mode B: groksessions.Info + RunForeground / exec mock
doctest <- stdout/stderr + recorders + returned error
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/agent/grok/fork` will export `Main`
  and `Options` (RED until then).
- No process-global env/cwd; no live iTerm2; no real `grok` TUI.
- Session lookup uses `opts.GrokHome` (fixture `{temp}/.grok`).
- `GROK_HOME` in `opts.Env` is forwarded on the Mode A follow-up only when set.
- Session-scoped mock binary cache:
  `$TMPDIR/grok-fork-doctest-<d.DOCTEST_SESSION_ID>/`
  `llm-mock-run-grok` + `binaries.ready` (file lock). Only the exec leaf builds.

## Steps

1. Root `Setup` allocates temp grok home + workspace and default pids.
2. Grouping / leaf `Setup` seeds sessions, procs, args.
3. `Run` calls `fork.Main` with recorders (or real exec when `ExecMock`).
4. Leaf `Assert` checks stdout/stderr/error/recorders.

## Context

- Default fixture session: `019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa`
- Nested main session: `019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb`
- Alternate `--pid` session: `019f283b-cccc-7ccc-cccc-cccccccccccc`
- Default chain: grok `4242` → bash `5000` → start `6000`
- Grok cmdline includes ignored `--resume` / `--session-id` decoys.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

const (
	fixtureSessionID     = "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa"
	fixtureMainSessionID = "019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb"
	fixtureAltSessionID  = "019f283b-cccc-7ccc-cccc-cccccccccccc"
	wrongResumeSessionID = "00000000-0000-0000-0000-000000000000"
	wrongFlagSessionID   = "11111111-1111-1111-1111-111111111111"

	pidGrok     = 4242
	pidBash     = 5000
	pidStart    = 6000
	pidMainGrok = 3000
	pidAltGrok  = 7000
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.TempDir == "" {
		req.TempDir = t.TempDir()
	}
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, ".grok")
	}
	if req.Workspace == "" {
		req.Workspace = filepath.Join(req.TempDir, "ws")
	}
	if err := os.MkdirAll(req.GrokHome, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.Workspace, 0o755); err != nil {
		return err
	}
	req.Workspace = absPath(t, req.Workspace)
	if req.PID == 0 {
		req.PID = pidStart
	}
	if req.Executable == "" {
		req.Executable = filepath.Join(req.TempDir, "grok-fork")
	}
	if req.GrokBin == "" {
		req.GrokBin = filepath.Join(req.TempDir, "llm-mock-run-grok")
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	if req.Env == nil {
		req.Env = []string{}
	}
	if req.Args == nil {
		req.Args = []string{}
	}
	return nil
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	if real, e := filepath.EvalSymlinks(abs); e == nil {
		return real
	}
	return abs
}

func encodeCWD(t *testing.T, cwd string) string {
	t.Helper()
	return url.PathEscape(absPath(t, cwd))
}

func grokCmdWithIgnoredFlags() string {
	return "/usr/local/bin/grok --resume " + wrongResumeSessionID + " --session-id " + wrongFlagSessionID
}

func defaultAncestorChain() []FixtureProc {
	return []FixtureProc{
		{PID: pidGrok, PPID: 1, Cmd: grokCmdWithIgnoredFlags()},
		{PID: pidBash, PPID: pidGrok, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
	}
}

func seedSession(t *testing.T, grokHome, sessionID, cwd string) string {
	t.Helper()
	absCWD := cwd
	if cwd != "" {
		absCWD = absPath(t, cwd)
	}
	key := "empty-cwd"
	if absCWD != "" {
		key = encodeCWD(t, absCWD)
	}
	dir := filepath.Join(grokHome, "sessions", key, sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": absCWD,
		},
		"generated_title": "grok-fork fixture",
		"created_at":      "2026-08-01T10:00:00.000Z",
		"updated_at":      "2026-08-01T11:00:00.000Z",
		"last_active_at":  "2026-08-01T11:00:00.000Z",
		"num_messages":    1,
		"num_chat_messages": 1,
	}
	raw, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), raw, 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}
	return dir
}

func seedModeA(t *testing.T, req *Request) {
	t.Helper()
	if len(req.Procs) == 0 {
		req.Procs = defaultAncestorChain()
	}
	dir := seedSession(t, req.GrokHome, fixtureSessionID, req.Workspace)
	req.OpenFiles = map[int][]string{
		pidGrok: {filepath.Join(dir, "events.jsonl")},
	}
}

func seedModeB(t *testing.T, req *Request) {
	t.Helper()
	seedSession(t, req.GrokHome, fixtureSessionID, req.Workspace)
}

func lsofPath(sessionDir string) string {
	return filepath.Join(sessionDir, "events.jsonl")
}

func assertMainOK(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Err != nil {
		t.Fatalf("unexpected Main error: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d, want 0", resp.ExitCode)
	}
}

func assertMainErr(t *testing.T, resp *Response, substrs ...string) {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Err == nil {
		t.Fatalf("expected Main error, got nil\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("ExitCode=%d, want 1", resp.ExitCode)
	}
	msg := resp.Err.Error()
	if strings.HasPrefix(msg, "Error:") {
		t.Fatalf("Main error must not include Error: prefix (thin cmd does that): %q", msg)
	}
	for _, s := range substrs {
		if !strings.Contains(msg, s) {
			t.Fatalf("error %q missing %q", msg, s)
		}
	}
}

func assertNoOpen(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.OpenCalls) != 0 {
		t.Fatalf("OpenInNewTerminal called %d times, want 0: %+v", len(resp.OpenCalls), resp.OpenCalls)
	}
}

func assertNoForeground(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.ForegroundCalls) != 0 {
		t.Fatalf("RunForeground called %d times, want 0: %+v", len(resp.ForegroundCalls), resp.ForegroundCalls)
	}
}

func assertOneOpen(t *testing.T, resp *Response) OpenCall {
	t.Helper()
	if len(resp.OpenCalls) != 1 {
		t.Fatalf("OpenInNewTerminal called %d times, want 1: %+v", len(resp.OpenCalls), resp.OpenCalls)
	}
	return resp.OpenCalls[0]
}

func assertOneForeground(t *testing.T, resp *Response) ForegroundCall {
	t.Helper()
	if len(resp.ForegroundCalls) != 1 {
		t.Fatalf("RunForeground called %d times, want 1: %+v", len(resp.ForegroundCalls), resp.ForegroundCalls)
	}
	return resp.ForegroundCalls[0]
}

func assertNoANSI(t *testing.T, s, label string) {
	t.Helper()
	if strings.Contains(s, "\x1b[") {
		t.Fatalf("%s has ANSI, want none:\n%s", label, s)
	}
}

func assertHasANSI(t *testing.T, s, label string) {
	t.Helper()
	if !strings.Contains(s, "\x1b[") {
		t.Fatalf("%s missing ANSI:\n%s", label, s)
	}
}

func v3Exact(lines ...string) string {
	var b strings.Builder
	b.WriteString("---\nversion: 3\n---\n")
	for _, line := range lines {
		b.WriteString(regexp.QuoteMeta(line))
		b.WriteByte('\n')
	}
	return b.String()
}

func assertStdoutExact(t *testing.T, got string, lines ...string) {
	t.Helper()
	assert.Output(t, got, v3Exact(lines...))
}

func modeASuccessLine(sessionID string) string {
	return "Opened new window; launching grok-fork --session-id " + sessionID
}

func followUpSession(executable, sessionID string) string {
	return shell.ShellQuote(executable) + " --session-id " + sessionID
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "grok-fork-doctest-"+d.DOCTEST_SESSION_ID)
}

func withFileLock(lockPath string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func moduleRoot(d *session.Doctest) (string, error) {
	start := d.DOCTEST_ROOT
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "agent", "llm", "llm-mock", "llm-mock-run-grok")); err == nil {
				return dir, nil
			}
		}
		if filepath.Dir(dir) == dir {
			return "", os.ErrNotExist
		}
	}
}

func buildMockGrokOnce(t *testing.T, d *session.Doctest) string {
	t.Helper()
	root, err := moduleRoot(d)
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	cache := sessionCacheDir(d)
	binPath := filepath.Join(cache, "llm-mock-run-grok")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err = withFileLock(lock, func() error {
		if _, e := os.Stat(ready); e == nil {
			if _, e := os.Stat(binPath); e == nil {
				return nil
			}
		}
		if err := os.MkdirAll(cache, 0o755); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./agent/llm/llm-mock/llm-mock-run-grok")
		cmd.Dir = root
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go build llm-mock-run-grok: %w\n%s", err, be.String())
		}
		return os.WriteFile(ready, []byte("ok\n"), 0o644)
	})
	if err != nil {
		t.Fatalf("build llm-mock-run-grok: %v", err)
	}
	return binPath
}

func helpMentions(t *testing.T, stdout string) {
	t.Helper()
	for _, tok := range []string{"--session-id", "--dir", "--pid", "--dry-run", "--color", "--no-color"} {
		if !strings.Contains(stdout, tok) {
			t.Fatalf("help missing %q:\n%s", tok, stdout)
		}
	}
	if strings.Contains(stdout, "--new-terminal") {
		t.Fatalf("help must not mention --new-terminal:\n%s", stdout)
	}
	if strings.Contains(stdout, " -n") || strings.Contains(stdout, "-n,") || strings.HasPrefix(strings.TrimSpace(stdout), "-n") {
		// allow the letter n inside other words; reject a -n flag token
		for _, line := range strings.Split(stdout, "\n") {
			fields := strings.Fields(line)
			for _, f := range fields {
				if f == "-n" || strings.HasPrefix(f, "-n,") {
					t.Fatalf("help must not mention -n:\n%s", stdout)
				}
			}
		}
	}
}
```
