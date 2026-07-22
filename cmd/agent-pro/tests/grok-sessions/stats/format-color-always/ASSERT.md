## Expected

- Stats succeeds; tool errors and Sources available for coloring.
- Output contains at least one ANSI CSI sequence (`\x1b[`).
- Prefer evidence of green (source yes) and/or red (errors) and/or dim when present.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	out := resp.Output
	if out == "" {
		t.Fatal("empty output")
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("ColorMode always expected ANSI CSI sequences, got:\n%s", out)
	}
	// Soft checks: green (32) for source yes and/or red (31) for errors/ERROR column,
	// and/or dim (2). At least one of these SGR families should appear.
	hasGreen := strings.Contains(out, "\x1b[32m") || strings.Contains(out, ";32m")
	hasRed := strings.Contains(out, "\x1b[31m") || strings.Contains(out, ";31m")
	hasDim := strings.Contains(out, "\x1b[2m") || strings.Contains(out, ";2m")
	if !hasGreen && !hasRed && !hasDim {
		t.Fatalf("expected green, red, or dim SGR in colored stats output:\n%q", out)
	}
	assertContains(t, out, "Sources")
}
```
