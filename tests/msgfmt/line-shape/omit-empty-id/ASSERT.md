## Expected

```text
Chat history (1 message):
[bob] : ping
```

- Does not contain `message_id=`
- `LastMessageID=""` (newest input id is empty)

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	want := "Chat history (1 message):\n[bob] : ping\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertNotContains(t, resp.Text, "message_id=")
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "")
	assertFormatEqualsDetail(t, resp)
}
```
