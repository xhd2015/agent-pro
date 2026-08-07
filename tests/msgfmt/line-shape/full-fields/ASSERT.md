## Expected

Exact text (trailing newline after the message line):

```text
Chat history (1 message):
message_id=m1  [alice] : hello
```

- `Shown=1`, `SourceCount=1`, `OldestDropped=0`, `BodiesTruncated=0`
- `LastMessageID="m1"`
- `Format` == `Detail.Text`

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	want := "Chat history (1 message):\nmessage_id=m1  [alice] : hello\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertEqualInt(t, "Shown", resp.Detail.Shown, 1)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 1)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 0)
	assertEqualInt(t, "BodiesTruncated", resp.Detail.BodiesTruncated, 0)
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "m1")
	assertFormatEqualsDetail(t, resp)
}
```
