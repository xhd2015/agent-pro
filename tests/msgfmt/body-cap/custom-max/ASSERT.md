## Expected

```text
Chat history (1 message):
message_id=m1  [alice] : abcd…
```

- Body rune count is 5.
- `BodiesTruncated=1`.

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	want := "Chat history (1 message):\nmessage_id=m1  [alice] : abcd" + truncationMarker + "\n"
	assertEqualString(t, "Text", resp.Text, want)
	lines := messageLines(resp.Text)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	if n := bodyRuneCountOfLine(lines[0]); n != 5 {
		t.Fatalf("body runes=%d, want 5", n)
	}
	assertEqualInt(t, "BodiesTruncated", resp.Detail.BodiesTruncated, 1)
	assertFormatEqualsDetail(t, resp)
}
```
