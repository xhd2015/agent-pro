## Expected

- `msgfmt.DefaultMaxPerMessageRunes == 1000`.
- Output contains the full 1000-`a` body (no `…`).
- Message line body rune count is 1000.
- `BodiesTruncated=0`.

## Errors

- None from `Run`.

```go
import (
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoError(t, err)
	assertResp(t, resp)
	if msgfmt.DefaultMaxPerMessageRunes != 1000 {
		t.Fatalf("DefaultMaxPerMessageRunes=%d, want 1000", msgfmt.DefaultMaxPerMessageRunes)
	}
	body := runeRepeat("a", 1000)
	assertContains(t, resp.Text, body)
	assertNotContains(t, resp.Text, truncationMarker)
	lines := messageLines(resp.Text)
	if len(lines) != 1 {
		t.Fatalf("want 1 message line, got %d in %q", len(lines), resp.Text)
	}
	if n := bodyRuneCountOfLine(lines[0]); n != 1000 {
		t.Fatalf("body runes=%d, want 1000 (line %q)", n, lines[0])
	}
	if utf8.RuneCountInString(body) != 1000 {
		t.Fatal("fixture body must be 1000 runes")
	}
	assertEqualInt(t, "BodiesTruncated", resp.Detail.BodiesTruncated, 0)
	assertEqualInt(t, "Shown", resp.Detail.Shown, 1)
	assertFormatEqualsDetail(t, resp)
	_ = req
}
```
