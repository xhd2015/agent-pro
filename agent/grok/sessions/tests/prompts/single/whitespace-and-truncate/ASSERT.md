## Expected

- No error.
- Structured `UserPrompts[0].Text` retains the raw multi-whitespace form (or
  at least contains both `hello` and a long `xxx` run — implementer may
  normalize on ingest; format is the hard contract).
- Formatted Output:
  - collapses internal whitespace to single spaces (no raw `\t` / raw newlines
    inside the prompt body line)
  - soft-truncates the body portion to about **200 runes** ending with `…`
  - trailing newline
  - no `👤`

## Errors

- None.

```go
import (
	"strings"
	"testing"
	"unicode/utf8"
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
	// Allow timestamp brackets; body should not include literal tab.
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
	// Truncation: ellipsis present and total body not huge.
	if !strings.Contains(out, "…") && !strings.Contains(out, "...") {
		t.Fatalf("expected truncation ellipsis in output:\n%s", out)
	}
	// Soft cap ~200 runes for the prompt text portion; whole line may be a bit longer.
	// Fail if body still has ~220 raw x's untruncated.
	if strings.Contains(out, longPromptRunes(220)) {
		t.Fatalf("expected long body to be truncated, still full 220 x run")
	}
	// Formatted output should be well under raw length.
	if utf8.RuneCountInString(out) > 280 {
		t.Fatalf("formatted output still very long (%d runes):\n%s", runeLen(out), out)
	}
	assertContains(t, out, "hello")
	assertContains(t, out, "world")
}
```
