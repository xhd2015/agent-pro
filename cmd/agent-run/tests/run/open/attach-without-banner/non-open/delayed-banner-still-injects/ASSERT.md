## Expected

- Combined output must **not** contain `banner not detected` (delayed banner still within wait).
- Probe file records `STDIN=hi` — prompt was injected **after** the delayed banner
  (if inject raced before banner, `read` would miss and probe would be empty/missing).
- Session id may appear on stderr (`grok-tty: <id>`); discovery/stream failures after
  inject are out of scope for this policy leaf.

## Errors

- Must not fail with grok TUI banner timeout.

## Exit Code

0 preferred; non-zero allowed only when banner wait clearly succeeded (probe has
`STDIN=hi` and no banner error) — e.g. late discovery cancel after inject.

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
		t.Fatalf("read inject probe %s: %v (banner wait/inject may have failed)\nstderr:\n%s",
			probePath, rerr, resp.Stderr)
	}
	probe := string(data)
	want := "STDIN=" + req.Prompt
	if !strings.Contains(probe, want) {
		t.Fatalf("expected inject after delayed banner (%q); probe:\n%s\nstdout:\n%s\nstderr:\n%s",
			want, probe, resp.Stdout, resp.Stderr)
	}

	// Prefer clean success; allow non-zero only after inject proof (discovery flake).
	if resp.ExitCode != 0 {
		// Still GREEN for banner-policy if inject landed; soft-check only.
		if !strings.Contains(strings.ToLower(combined), "discover") &&
			!strings.Contains(strings.ToLower(combined), "resolve session") {
			t.Fatalf("exit=%d after successful inject; unexpected failure:\n%s", resp.ExitCode, resp.Stderr)
		}
	}
}
```
