## Expected

- Stats succeeds with tool failures / errors > 0 and Sources present.
- Output has **no** ANSI CSI sequences (`\x1b`).

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
	if resp.Stats.ToolFailed < 1 && resp.Stats.Errors < 1 {
		// Fixture should include failures so "never" is meaningful vs always.
		t.Fatalf("fixture expected tool failures or errors, got failed=%d errors=%d",
			resp.Stats.ToolFailed, resp.Stats.Errors)
	}
	out := resp.Output
	if out == "" {
		t.Fatal("empty output")
	}
	if strings.Contains(out, "\x1b") {
		t.Fatalf("ColorMode never must not emit ANSI, got:\n%q", out)
	}
	assertContains(t, out, "Sources")
	assertContains(t, out, "yes")
}
```
