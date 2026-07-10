## Expected

When a distinctive process holds the registry flock, `tty-watch run bash -c 'echo yes'`
must fail promptly with **richer lock-busy diagnostics** on stderr (or the CLI error
path), not hang and not emit only the short misleading one-liner.

Fixed content contract (see requirement design):

1. **Non-zero** exit; **not** harness `TimedOut`.
2. Summary containing stable prefix **`registry lock busy`**.
3. **Lock path** — absolute path to `{TTY_WATCH_HOME}/registry/.lock` (harness `resp.LockPath`).
4. **Holders section** — keyword `holder`/`holders` plus the fixture **holder PID** and
   distinctive **command marker** (`resp.LockHolderMarker` from lockholder argv).
5. **Process tree** markers — at least one of: `process tree`, `parent`, `children`, or `└`
   (ancestors and/or children of the holder).
6. Must **not** be solely the short message that only blames
   `another tty-watch run may be in progress` without holders/tree lines.
7. **stdout** must not print PTY/command output (`yes`) or `session-` ids when lock fails.
8. Multi-line stderr ends with trailing newline `\n` (CLI convention).

Reference shape (dynamic PIDs/paths; assert structure + fixture values, not full template):

```text
tty-watch: registry lock busy after 1.5s
  lock:  /…/registry/.lock

  holders (exclusive flock):
    PID    PPID   COMMAND
    <pid>  …      …/lockholder … <marker>

  process tree (holder → root):
    …

  children of holder (depth ≤2):
    …

  what to do:
    - …
```

## Errors

- Non-zero exit (lock busy).
- Harness timeout ⇒ fail (`TimedOut`).

## Exit Code

Non-zero.

```go
import (
	"strconv"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.TimedOut {
		t.Fatalf("tty-watch run hung while registry .lock held (timed out after %s); expected prompt lock diagnostics, output %q",
			resp.Elapsed, resp.Combined)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when registry flock held, got 0\nstdout:%q\nstderr:%q",
			resp.Stdout, resp.Stderr)
	}

	// Diagnostics go to stderr; Combined is fallback if a path merges streams.
	text := resp.Stderr
	if strings.TrimSpace(text) == "" {
		text = resp.Combined
	}
	if strings.TrimSpace(text) == "" {
		text = resp.Stdout
	}
	lower := strings.ToLower(text)

	// 1) Stable summary prefix
	if !strings.Contains(lower, "registry lock busy") {
		t.Fatalf("expected summary containing %q, got %q", "registry lock busy", text)
	}

	// 2) Absolute lock path from fixture
	if resp.LockPath == "" {
		t.Fatal("harness did not set resp.LockPath")
	}
	if !strings.Contains(text, resp.LockPath) {
		t.Fatalf("expected absolute lock path %q in diagnostics, got %q", resp.LockPath, text)
	}

	// 3) Holder PID + distinctive command marker
	if resp.LockHolderPID <= 0 {
		t.Fatalf("harness did not set LockHolderPID, got %d", resp.LockHolderPID)
	}
	pidStr := strconv.Itoa(resp.LockHolderPID)
	if !strings.Contains(text, pidStr) {
		t.Fatalf("expected holder PID %s in diagnostics, got %q", pidStr, text)
	}
	if resp.LockHolderMarker == "" {
		t.Fatal("harness did not set LockHolderMarker")
	}
	if !strings.Contains(text, resp.LockHolderMarker) {
		t.Fatalf("expected holder command marker %q in diagnostics, got %q", resp.LockHolderMarker, text)
	}

	// 4) Holders section
	if !strings.Contains(lower, "holder") {
		t.Fatalf("expected holders section (keyword holder/holders) in diagnostics, got %q", text)
	}

	// 5) Process tree markers (ancestors and/or children)
	hasTree := strings.Contains(lower, "process tree") ||
		strings.Contains(lower, "parent") ||
		strings.Contains(lower, "children") ||
		strings.Contains(text, "└")
	if !hasTree {
		t.Fatalf("expected process tree markers (process tree / parent / children / └), got %q", text)
	}

	// 6) Reject short one-liner that only blames "another tty-watch run"
	// without multi-line holders/tree diagnostics.
	trimmed := strings.TrimSpace(text)
	if !strings.Contains(trimmed, "\n") &&
		strings.Contains(lower, "another tty-watch run") {
		t.Fatalf("still the short misleading one-liner without holders/tree diagnostics: %q", text)
	}

	// 7) stdout silent on lock failure (no PTY output / session-id)
	if strings.Contains(resp.Stdout, "yes") {
		t.Fatalf("stdout must not show command/PTY output when lock busy: %q", resp.Stdout)
	}
	for _, line := range strings.Split(resp.Stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "session-") || strings.HasPrefix(trim, "session-id:") {
			t.Fatalf("stdout must not print session-id when lock busy: %q", resp.Stdout)
		}
	}

	// 8) Trailing newline on multi-line stderr (CLI convention)
	if resp.Stderr != "" && !strings.HasSuffix(resp.Stderr, "\n") {
		t.Fatalf("stderr diagnostics must end with trailing newline, got %q", resp.Stderr)
	}
}
```
