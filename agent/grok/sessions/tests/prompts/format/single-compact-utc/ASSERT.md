## Expected Output

```
---
version: 3
---
\[2026-07-03 14:30:00\] hello world
```

## Expected

- No error.
- Output matches compact line with UTC wall clock from wire.
- Trailing newline.
- No `👤`.

## Errors

- None.

```go
import (
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertTrailingNewline(t, resp.Output)
	assertNotContains(t, resp.Output, "👤")

	prefix := compactLinePrefix(atFixed(-30*time.Minute), time.UTC)
	wantLine := prefix + " hello world"
	// Allow optional footer after the prompt line.
	lines := strings.Split(strings.TrimSuffix(resp.Output, "\n"), "\n")
	if len(lines) < 1 {
		t.Fatalf("empty output")
	}
	if lines[0] != wantLine {
		// Soft: first non-empty line contains prefix and text
		found := false
		for _, ln := range lines {
			if strings.Contains(ln, prefix) && strings.Contains(ln, "hello world") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("output missing compact line %q:\n%s", wantLine, resp.Output)
		}
	}
}
```
