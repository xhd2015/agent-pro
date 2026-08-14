## Expected

- `resp.InputBox` is `occupied`.
- Last `›` line is `› EXP-DRAFT-NOTE-42` and does **not** contain ` medium · `.
- A later line contains the footer (moved off the composer line).

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	raw, readErr := os.ReadFile(filepath.Join(req.TestdataDir, fixtureCodexOccupiedSingle))
	if readErr != nil {
		t.Fatal(readErr)
	}
	text := string(raw)
	idx := strings.LastIndex(text, "›")
	if idx < 0 {
		t.Fatal("fixture must contain ›")
	}
	line := text[idx:]
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}
	if strings.Contains(line, " medium · ") {
		t.Fatalf("last › line must not contain footer glue, got %q", line)
	}
	if !strings.Contains(line, "EXP-DRAFT-NOTE-42") {
		t.Fatalf("last › line must be the draft, got %q", line)
	}
	if !strings.Contains(text, " medium · ") {
		t.Fatal("fixture must still include the footer on a later line")
	}
	assertInputBox(t, resp, err, "occupied")
}
```
