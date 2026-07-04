## Expected

- Exit code 0.
- Snapshot shows **one** latest MCP status line (no stacked boot fragments).
- No kitty/CSI leak fragments like `[<` or `;23M`.
- No smeared long lines (max line length 120).
- No CSI/OSC/C0 escape leaks.

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
	if count := strings.Count(text, "Starting MCP servers"); count > 1 {
		t.Fatalf("snapshot smeared %d MCP boot lines (want 1 latest), got %q", count, text)
	}
	if strings.Contains(text, "[<") || strings.Contains(text, ";23M") {
		t.Fatalf("snapshot leaked kitty/CSI fragments, got %q", text)
	}
	if resp.ContainsEscape {
		t.Fatalf("snapshot still contains escape sequences: %q", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if len(line) > 120 {
			t.Fatalf("snapshot smeared long line (%d chars): %q", len(line), line)
		}
	}
}
```