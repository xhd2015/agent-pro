## Expected

- No error.
- Formatted body is soft-capped: prefix of **200** `x` content runes + `…`.
- Full 220-x run is **not** present.
- Unicode ellipsis `…` (U+2026) present; trailing newline.

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

	if strings.Contains(out, longPromptRunes(220)) {
		t.Fatalf("MaxBody=200 must not print full 220 x run:\n%s", out)
	}
	// Body after "] " should be 200 x's + …
	body := out
	if i := strings.Index(out, "] "); i >= 0 {
		body = strings.TrimSuffix(out[i+2:], "\n")
		// Allow multi-line only for footer; take first line as prompt body.
		if j := strings.Index(body, "\n"); j >= 0 {
			body = body[:j]
		}
	}
	wantPrefix := longPromptRunes(200)
	if !strings.HasPrefix(body, wantPrefix) {
		t.Fatalf("expected body to start with 200 x runes, got %q (runes=%d)", body, utf8.RuneCountInString(body))
	}
	if !strings.HasSuffix(body, "…") {
		t.Fatalf("expected body to end with … (U+2026), got %q", body)
	}
	// Content runes before ellipsis = 200; ellipsis is outside N.
	content := strings.TrimSuffix(body, "…")
	if got := utf8.RuneCountInString(content); got != 200 {
		t.Fatalf("content runes before … = %d want 200 (body=%q)", got, body)
	}
}
```
