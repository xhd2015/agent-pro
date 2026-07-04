## Expected

- Exit code 0.
- Snapshot shows the **latest** codex-like screen only: `OpenAI Codex` appears once.
- Snapshot contains `Improve documentation` once (current prompt line).
- No smeared long lines from stacked redraws (max line length 120).
- No orphaned true-color SGR fragments like `[38;2;`.
- No CSI/OSC/C0 escape leaks in output.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("snapshot exit %d stdout %q stderr %q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	text := resp.SnapshotText
	if text == "" {
		text = resp.Stdout
	}
	if count := strings.Count(text, "OpenAI Codex"); count != 1 {
		t.Fatalf("snapshot stacked %d UI headers (want 1 latest screen), got %q", count, text)
	}
	if count := strings.Count(text, "Improve documentation"); count != 1 {
		t.Fatalf("snapshot stacked %d prompt lines (want 1), got %q", count, text)
	}
	if strings.Contains(text, "[38;2;") {
		t.Fatalf("snapshot leaked orphaned true-color SGR fragments, got %q", text)
	}
	if resp.ContainsEscape {
		t.Fatalf("snapshot still contains escape sequences: %q", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if len(line) > 120 {
			t.Fatalf("snapshot smeared long line (%d chars) from stacked redraws: %q", len(line), line)
		}
	}
}
```