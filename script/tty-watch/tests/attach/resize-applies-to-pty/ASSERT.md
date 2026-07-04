## Expected

- Snapshot exit code 0.
- Attach handshake reports `attach_role` attacher (not observer) while screen writer is connected.
- Wide marker line appears once on a single row (not 80-column wrapped).
- Snapshot reflects resized session width (cols=100).

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
	if !strings.Contains(resp.AttachOutput, `"attach_role":"attacher"`) {
		t.Fatalf("attach resize requires attach_mode=attach attacher role, handshake %q", resp.AttachOutput)
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