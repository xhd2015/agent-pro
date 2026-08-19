## Expected

- `ready=false`, `state=busy`, reason mentions working.
- Fixture contains live `• Working` / `esc to interrupt` **and** a later placeholder `›`.
- Fixture must **not** contain settled `Worked for` (that is F8, post-turn idle).

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
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureWorkingAbovePlaceholderBusy))
	if err != nil {
		t.Fatal(err)
	}
	s := string(text)
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "working") {
		t.Fatalf("fixture must contain live Working text")
	}
	if !strings.Contains(lower, "esc to interrupt") {
		t.Fatalf("fixture must contain esc to interrupt")
	}
	if strings.Contains(lower, "worked for") {
		t.Fatalf("fixture must not be a settled Worked-for footer (that is F8 idle)")
	}
	if !strings.Contains(s, "›") && !strings.Contains(s, "\u203a") {
		t.Fatalf("fixture must contain composer ›")
	}
	assertWritable(t, "working-above-placeholder-prompt-busy", resp.Status, false, "busy", "working")
}
```
