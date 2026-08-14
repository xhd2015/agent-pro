## Expected

- `resp.InputBox` is `occupied`.
- Last `›` line contains `LINE1-DRAFT` and does not contain ` medium · `.
- Following line contains `LINE2-DRAFT` and may glue to the footer.

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
	raw, readErr := os.ReadFile(filepath.Join(req.TestdataDir, fixtureCodexOccupiedMultiline))
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
	if !strings.Contains(line, "LINE1-DRAFT") {
		t.Fatalf("last › line must be LINE1-DRAFT, got %q", line)
	}
	if !strings.Contains(text, "LINE2-DRAFT") {
		t.Fatal("fixture must include wrapped LINE2-DRAFT")
	}
	assertInputBox(t, resp, err, "occupied")
}
```
