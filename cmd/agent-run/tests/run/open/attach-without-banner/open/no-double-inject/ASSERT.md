---
label: e2e
---

## Expected

- Exit code 0; no `banner not detected`.
- Post-attach `grok-tty: <id>` once on stderr.
- Probe file records `PROMPT_ARG=once-only` (argv path).
- Probe `STDIN_COUNT=0` — open must not re-inject the new-session prompt via PTY.
- If `STDIN=` is present with the same prompt text, fail (double-submit).

## Side Effects

- Probe file under leaf temp dir (`tui-probe.txt`).

## Exit Code

0

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
		t.Fatalf("--open must not hard-fail on missing banner for argv prompt:\n%s", resp.Stderr)
	}
	assertSuccess(t, resp)

	if _, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty"); !ok {
		t.Fatalf("missing post-attach grok-tty session id on stderr:\n%s", resp.Stderr)
	}

	probePath := filepath.Join(req.TempDir, "tui-probe.txt")
	// Prefer path written by Setup if present.
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
	if !strings.Contains(probe, "PROMPT_ARG="+req.Prompt) && !strings.Contains(probe, "PROMPT_ARG="+req.Prompt+"\n") {
		// Also accept PROMPT_ARG as last argv token in ARGV= line.
		if !strings.Contains(probe, req.Prompt) {
			t.Fatalf("probe missing argv prompt %q:\n%s", req.Prompt, probe)
		}
	}
	if strings.Contains(probe, "STDIN_COUNT=1") || strings.Contains(probe, "STDIN=") {
		// Fail on any PTY inject — new-session open should not re-inject.
		t.Fatalf("new-session --open must not re-inject prompt via PTY; probe:\n%s\nstderr:\n%s",
			probe, resp.Stderr)
	}
	if !strings.Contains(probe, "STDIN_COUNT=0") {
		t.Fatalf("probe missing STDIN_COUNT=0 (did fake TUI write probe?):\n%s", probe)
	}
}
```
