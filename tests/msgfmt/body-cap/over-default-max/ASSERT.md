## Expected

- Body is `999×a` + `…` (U+2026); total body runes = 1000.
- Full original 1001-`a` string does **not** appear.
- Does not use ASCII `...` or `[truncated]`.
- `BodiesTruncated=1`.

Exact text:

```text
Chat history (1 message):
message_id=m1  [alice] : <999 a's>…
```

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	wantBody := runeRepeat("a", 999) + truncationMarker
	want := "Chat history (1 message):\nmessage_id=m1  [alice] : " + wantBody + "\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertNotContains(t, resp.Text, runeRepeat("a", 1001))
	assertNotContains(t, resp.Text, "...")
	assertNotContains(t, resp.Text, "[truncated]")
	assertContains(t, resp.Text, truncationMarker)
	lines := messageLines(resp.Text)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	if n := bodyRuneCountOfLine(lines[0]); n != 1000 {
		t.Fatalf("truncated body runes=%d, want 1000", n)
	}
	assertEqualInt(t, "BodiesTruncated", resp.Detail.BodiesTruncated, 1)
	assertFormatEqualsDetail(t, resp)
}
```
