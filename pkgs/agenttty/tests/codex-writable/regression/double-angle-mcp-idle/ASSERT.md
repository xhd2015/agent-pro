## Expected

- `ready=true`, `state=idle` (MCP incomplete does not block when main prompt is present).
- Fixture contains `MCP startup incomplete` and main chat `»` (U+00BB).
- Fixture **must not** contain legacy `›` (U+203A).

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
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureDoubleAngleMCP))
	if err != nil {
		t.Fatal(err)
	}
	s := string(text)
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "mcp startup incomplete") {
		t.Fatalf("fixture must contain MCP startup incomplete")
	}
	if !strings.Contains(s, "»") && !strings.Contains(s, "\u00bb") {
		t.Fatalf("fixture must contain double-angle prompt » (U+00BB)")
	}
	if strings.Contains(s, "›") || strings.Contains(s, "\u203a") {
		t.Fatalf("fixture must not contain legacy › (U+203A) — would mask the bug")
	}
	assertWritable(t, "double-angle-mcp-idle", resp.Status, true, "idle", "")
}
```
