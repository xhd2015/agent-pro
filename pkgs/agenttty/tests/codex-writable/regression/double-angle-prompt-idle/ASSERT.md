## Expected

- `ready=true`, `state=idle`.
- Fixture contains main chat `»` (U+00BB) and **must not** contain legacy `›` (U+203A).
- `DetectScreenStatus` is **not** `unknown` / empty — expect `banner` or `idle`
  (same class as ›-only idle fixtures without `response:` / `submitted:`).

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureDoubleAngleIdle))
	if err != nil {
		t.Fatal(err)
	}
	s := string(text)
	if !strings.Contains(s, "»") && !strings.Contains(s, "\u00bb") {
		t.Fatalf("fixture must contain double-angle prompt » (U+00BB)")
	}
	if strings.Contains(s, "›") || strings.Contains(s, "\u203a") {
		t.Fatalf("fixture must not contain legacy › (U+203A) — would mask the bug")
	}
	if !strings.Contains(s, "usage limit") {
		t.Fatalf("fixture must include usage-limit bullet chrome from incident")
	}
	assertWritable(t, "double-angle-prompt-idle", resp.Status, true, "idle", "")

	provider, ok := agenttty.Get("codex-tty")
	if !ok {
		t.Fatal("codex-tty provider not registered")
	}
	if provider.DetectScreenStatus == nil {
		t.Fatal("codex-tty DetectScreenStatus is nil")
	}
	screen := provider.DetectScreenStatus(text)
	if screen == "" || screen == "unknown" {
		t.Fatalf("DetectScreenStatus must not stay unknown for » idle chat (got %q)", screen)
	}
	if screen != "banner" && screen != "idle" {
		t.Fatalf("DetectScreenStatus got %q want banner|idle (match ›-only idle class)", screen)
	}
}
```
