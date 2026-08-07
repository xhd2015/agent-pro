## Expected

```text
Chat history (1 message):
message_id=m1  [alice] : 你好…
```

- Body is exactly three runes (`你`, `好`, `…`).
- Full original `你好世界` does not appear.
- `BodiesTruncated=1`.

## Errors

- None from `Run`.

```go
import (
	"testing"
	"unicode/utf8"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	wantBody := "你好" + truncationMarker
	want := "Chat history (1 message):\nmessage_id=m1  [alice] : " + wantBody + "\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertNotContains(t, resp.Text, "你好世界")
	if utf8.RuneCountInString(wantBody) != 3 {
		t.Fatalf("fixture wantBody runes=%d", utf8.RuneCountInString(wantBody))
	}
	lines := messageLines(resp.Text)
	if n := bodyRuneCountOfLine(lines[0]); n != 3 {
		t.Fatalf("body runes=%d, want 3", n)
	}
	assertEqualInt(t, "BodiesTruncated", resp.Detail.BodiesTruncated, 1)
	assertFormatEqualsDetail(t, resp)
}
```
