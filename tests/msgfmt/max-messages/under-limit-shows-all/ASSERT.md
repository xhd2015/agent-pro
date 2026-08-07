## Expected

Exact text:

```text
Chat history (showing 3 of 3):
message_id=m1  [a] : one
message_id=m2  [b] : two
message_id=m3  [c] : three
```

- `Shown=3`, `SourceCount=3`, `OldestDropped=0`
- `LastMessageID="m3"`

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	want := "" +
		"Chat history (showing 3 of 3):\n" +
		"message_id=m1  [a] : one\n" +
		"message_id=m2  [b] : two\n" +
		"message_id=m3  [c] : three\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertEqualInt(t, "Shown", resp.Detail.Shown, 3)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 3)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 0)
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "m3")
	assertFormatEqualsDetail(t, resp)
}
```
