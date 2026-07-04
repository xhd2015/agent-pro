## Expected

- Exit code 0.
- Snapshot contains the full wide marker line (`WIDE_LINE_MARKER_` + 75 chars) exactly once.
- The marker appears on a **single** line (no 80-column wrap split).
- No snapshot line exceeds 100 characters (marker length 92 fits within cols=100).

```go
import (
	"strings"
	"testing"
)

const wideLineMarkerPrefix = "WIDE_LINE_MARKER_"

func wideLineMarker() string {
	return wideLineMarkerPrefix + strings.Repeat("x", 75)
}

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
	marker := wideLineMarker()
	if strings.Count(text, marker) != 1 {
		t.Fatalf("snapshot wide marker count %d (want 1 contiguous line), got %q", strings.Count(text, marker), text)
	}
	foundOnSingleLine := false
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, marker) {
			foundOnSingleLine = true
		}
		if len(line) > 100 {
			t.Fatalf("snapshot line exceeds session width (%d > 100 cols): %q", len(line), line)
		}
	}
	if !foundOnSingleLine {
		t.Fatalf("wide marker split across lines (80-col wrap?), want single line %q in %q", marker, text)
	}
	if resp.ContainsEscape {
		t.Fatalf("snapshot still contains escape sequences: %q", text)
	}
}
```