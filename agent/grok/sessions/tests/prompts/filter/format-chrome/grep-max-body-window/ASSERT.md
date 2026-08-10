## Expected

- No error.
- Output contains `MATCH`.
- Body is windowed: **not** full 150 a's / 150 b's runs.
- Side ellipsis `…` present (cuts on both sides of a centered match).
- Window content (excluding side ellipses) is within MaxBody budget (~40 runes);
  whole body portion well under full 305 runes.
- Trailing newline.

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
	out := resp.Output
	assertTrailingNewline(t, out)
	assertContains(t, out, "MATCH")

	if strings.Contains(out, strings.Repeat("a", 150)) {
		t.Fatalf("MaxBody window must not keep full 150 leading a's:\n%s", out)
	}
	if strings.Contains(out, strings.Repeat("b", 150)) {
		t.Fatalf("MaxBody window must not keep full 150 trailing b's:\n%s", out)
	}
	if !strings.Contains(out, "…") {
		t.Fatalf("expected side ellipsis from window cut:\n%s", out)
	}

	body := out
	if i := strings.Index(out, "] "); i >= 0 {
		body = strings.TrimSuffix(out[i+2:], "\n")
		if j := strings.Index(body, "\n"); j >= 0 {
			body = body[:j]
		}
	}
	// Body should be near MaxBody budget (content ≤ 40 + up to 2 ellipsis runes).
	if n := utf8.RuneCountInString(body); n > 50 {
		t.Fatalf("windowed body still too long (%d runes, MaxBody=40): %q", n, body)
	}
	if n := utf8.RuneCountInString(body); n < 5 {
		t.Fatalf("windowed body unexpectedly tiny (%d runes): %q", n, body)
	}
}
```
