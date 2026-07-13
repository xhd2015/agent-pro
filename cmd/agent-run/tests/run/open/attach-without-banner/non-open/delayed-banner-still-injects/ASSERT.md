## Expected

- Combined output must **not** contain `banner not detected` (delayed banner still within wait).
- Probe records `PROMPT_ARG=hi` (or prompt on argv) — new-session prompt stays on argv.
- Probe `STDIN_COUNT=0` — must **not** re-inject after delayed banner (double-submit fix).
- Presence of `STDIN=hi` / `STDIN_COUNT=1` fails this leaf (old buggy policy).
- Session id may appear on stderr (`grok-tty: <id>`); discovery/stream failures after
  policy point are out of scope for this leaf when no-inject is proven.

## Errors

- Must not fail with grok TUI banner timeout.

## Exit Code

0 preferred; non-zero allowed only when banner wait clearly succeeded and probe
has `STDIN_COUNT=0` with no `STDIN=` — e.g. late discovery cancel.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	// Context deadline on the whole exec still fails the leaf.
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	combined := resp.Stdout + "\n" + resp.Stderr
	if hasBannerNotDetected(combined) {
		t.Fatalf("non-open delayed banner must still succeed; banner error:\n%s", resp.Stderr)
	}

	probePath := filepath.Join(req.TempDir, "non-open-inject-probe.txt")
	for _, e := range req.Env {
		if strings.HasPrefix(e, "DOCTEST_TUI_PROBE_PATH=") {
			probePath = strings.TrimPrefix(e, "DOCTEST_TUI_PROBE_PATH=")
			break
		}
	}
	data, rerr := os.ReadFile(probePath)
	if rerr != nil {
		t.Fatalf("read inject probe %s: %v (banner wait may have failed)\nstderr:\n%s",
			probePath, rerr, resp.Stderr)
	}
	probe := string(data)

	// New-session: prompt must still be on argv (do not remove argv placement).
	if !strings.Contains(probe, "PROMPT_ARG="+req.Prompt) && !strings.Contains(probe, req.Prompt) {
		t.Fatalf("probe missing argv prompt %q:\n%s", req.Prompt, probe)
	}

	// Double-submit fix: no PTY re-inject of new-session prompt after delayed banner.
	if strings.Contains(probe, "STDIN_COUNT=1") || strings.Contains(probe, "STDIN=") {
		t.Fatalf("new-session non-open must not re-inject after delayed banner; probe:\n%s\nstdout:\n%s\nstderr:\n%s",
			probe, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(probe, "STDIN_COUNT=0") {
		t.Fatalf("probe missing STDIN_COUNT=0 (did fake TUI write probe?):\n%s", probe)
	}

	// Prefer clean success; allow non-zero only after no-inject proof (discovery flake).
	if resp.ExitCode != 0 {
		if !strings.Contains(strings.ToLower(combined), "discover") &&
			!strings.Contains(strings.ToLower(combined), "resolve session") {
			t.Fatalf("exit=%d after no-inject proof; unexpected failure:\n%s", resp.ExitCode, resp.Stderr)
		}
	}
}
```
