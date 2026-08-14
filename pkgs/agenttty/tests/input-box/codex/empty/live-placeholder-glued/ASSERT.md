## Expected

- `resp.InputBox` is `empty`.
- Fixture last composer line contains `›` and ` medium · ` on that same line.
- Remainder after `›` is non-empty (placeholder text) — occupancy is the glue, not TrimSpace.

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
	raw, readErr := os.ReadFile(filepath.Join(req.TestdataDir, fixtureCodexEmptyGlued))
	if readErr != nil {
		t.Fatal(readErr)
	}
	text := string(raw)
	idx := strings.LastIndex(text, "›")
	if idx < 0 {
		idx = strings.LastIndex(text, "\u203a")
	}
	if idx < 0 {
		t.Fatal("fixture must contain ›")
	}
	line := text[idx:]
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}
	if !strings.Contains(line, " medium · ") {
		t.Fatalf("last › line must contain footer glue %q, got %q", " medium · ", line)
	}
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "›"), "\u203a"))
	if rest == "" {
		t.Fatal("live empty fixture must keep placeholder text after ›")
	}
	assertInputBox(t, resp, err, "empty")
}
```
