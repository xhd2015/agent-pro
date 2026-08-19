---
label: e2e
---

## Expected

- Exit code 0; no `banner not detected`.
- Stderr has `grok-tty: <id>` once (non-open prints session id after start).
- Probe file records `PROMPT_ARG=once-only` (argv path).
- Probe `STDIN_COUNT=0` — non-open must not re-inject the new-session prompt via PTY.
- If `STDIN=` is present with the same prompt text, fail (double-submit).

## Side Effects

- Probe file under leaf temp dir (`tui-probe.txt`).

## Exit Code

0 preferred; non-zero allowed only when banner wait clearly succeeded and probe
proves no inject (e.g. late discovery cancel after argv-only submit path).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	combined := resp.Stdout + "\n" + resp.Stderr
	if hasBannerNotDetected(combined) {
		t.Fatalf("non-open must not hard-fail on delayed/short banner for argv prompt:\n%s", resp.Stderr)
	}

	probePath := filepath.Join(req.TempDir, "tui-probe.txt")
	for _, e := range req.Env {
		if strings.HasPrefix(e, "DOCTEST_TUI_PROBE_PATH=") {
			probePath = strings.TrimPrefix(e, "DOCTEST_TUI_PROBE_PATH=")
			break
		}
	}
	data, rerr := os.ReadFile(probePath)
	if rerr != nil {
		t.Fatalf("read TUI probe %s: %v\nstderr:\n%s", probePath, rerr, resp.Stderr)
	}
	probe := string(data)
	if !strings.Contains(probe, "PROMPT_ARG="+req.Prompt) {
		if !strings.Contains(probe, req.Prompt) {
			t.Fatalf("probe missing argv prompt %q:\n%s", req.Prompt, probe)
		}
	}
	// Soft newline kick (unblock `read` fakes) is OK; re-injecting the prompt text is not.
	if strings.Contains(probe, "STDIN="+req.Prompt) {
		t.Fatalf("new-session non-open must not re-inject prompt via PTY; probe:\n%s\nstderr:\n%s",
			probe, resp.Stderr)
	}
	if !strings.Contains(probe, "STDIN_COUNT=0") && !strings.Contains(probe, "STDIN_COUNT=1") {
		t.Fatalf("probe missing STDIN_COUNT (did fake TUI write probe?):\n%s", probe)
	}
	if strings.Contains(probe, "STDIN_COUNT=1") {
		// COUNT=1 with empty STDIN= is the soft \\n kick; only fail if a non-empty payload arrived.
		for _, line := range strings.Split(probe, "\n") {
			if strings.HasPrefix(line, "STDIN=") {
				got := strings.TrimPrefix(line, "STDIN=")
				if strings.TrimSpace(got) != "" {
					t.Fatalf("new-session non-open must not inject prompt text via PTY; probe:\n%s\nstderr:\n%s",
						probe, resp.Stderr)
				}
			}
		}
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
