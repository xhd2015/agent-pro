## Expected

Exact text:

```text
Chat history (showing last 3 of 10):
message_id=m08  [u] : b08
message_id=m09  [u] : b09
message_id=m10  [u] : b10
```

- Does not contain `b01` or `message_id=m01`
- `Shown=3`, `SourceCount=10`, `OldestDropped=7`
- `LastMessageID="m10"`

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
		"Chat history (showing last 3 of 10):\n" +
		"message_id=m08  [u] : b08\n" +
		"message_id=m09  [u] : b09\n" +
		"message_id=m10  [u] : b10\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertNotContains(t, resp.Text, "b01")
	assertNotContains(t, resp.Text, "message_id=m01")
	assertEqualInt(t, "Shown", resp.Detail.Shown, 3)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 10)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 7)
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "m10")
	assertFormatEqualsDetail(t, resp)
}
```
