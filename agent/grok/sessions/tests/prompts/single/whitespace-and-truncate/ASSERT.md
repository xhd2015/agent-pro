## Expected

- No error.
- Structured `UserPrompts[0].Text` retains the raw multi-whitespace form (or
  at least contains both `hello` and a long `xxx` run — implementer may
  normalize on ingest; format is the hard contract).
- Formatted Output:
  - collapses internal whitespace to single spaces (no raw `\t` / raw newlines
    inside the prompt body line)
  - **full** body by default: the collapsed `220` x-run is present
  - **no** body-cap ellipsis `…` (U+2026) from length truncation
  - trailing newline
  - no `👤`

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertPromptCount(t, resp.Single, 1)

	out := resp.Output
	assertTrailingNewline(t, out)
	assertNotContains(t, out, "👤")
	// Collapsed: should not keep tab as-is in the body line content after the bracket.
	body := out
	if i := strings.Index(out, "] "); i >= 0 {
		body = out[i+2:]
	}
	if strings.Contains(body, "\t") {
		t.Fatalf("formatted body still has tab after collapse:\n%s", out)
	}
	if strings.Contains(strings.TrimRight(body, "\n"), "\n") {
		t.Fatalf("formatted body is multi-line (expected single compact line):\n%q", out)
	}
	// Default full body: entire 220-x run must appear after collapse.
	xs := longPromptRunes(220)
	if !strings.Contains(out, xs) {
		t.Fatalf("expected full collapsed body with 220 x runes (no soft-cap):\n%s", out)
	}
	// No body-cap ellipsis when MaxBody is unset.
	if strings.Contains(out, "…") {
		t.Fatalf("default full body must not soft-cap with …:\n%s", out)
	}
	assertContains(t, out, "hello")
	assertContains(t, out, "world")
}
```
