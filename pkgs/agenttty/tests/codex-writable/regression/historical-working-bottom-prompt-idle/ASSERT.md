## Expected

- `ready=true`, `state=idle` (desired product: bottom settled `›` wins over historical Working).
- Fixture contains historical busy markers **and** main chat `›` / U+203A.
- Fixture contains settled footer `Worked for` so this is post-turn, not live working.

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
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureHistoricalWorkingBottomPromptIdle))
	if err != nil {
		t.Fatal(err)
	}
	s := string(text)
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "working") {
		t.Fatalf("fixture must contain historical Working text")
	}
	if !strings.Contains(lower, "esc to interrupt") {
		t.Fatalf("fixture must contain historical esc to interrupt")
	}
	if !strings.Contains(s, "•") && !strings.Contains(s, "\u2022") {
		t.Fatalf("fixture must contain bullet • used by the busy rule")
	}
	if !strings.Contains(s, "›") && !strings.Contains(s, "\u203a") {
		t.Fatalf("fixture must contain main chat › (U+203A)")
	}
	if !strings.Contains(lower, "worked for") {
		t.Fatalf("fixture must include settled Worked for footer (post-turn)")
	}
	// Desired: sendable/idle after turn complete despite historical busy chrome.
	// RED on current checkCodexWritable (full-scrollback •+working rule).
	assertWritable(t, "historical-working-bottom-prompt-idle", resp.Status, true, "idle", "")
}
```
