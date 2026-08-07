## Expected

```text
Chat history (1 message):
message_id=m9 : solo
```

- Does not contain `[` or `]`

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	want := "Chat history (1 message):\nmessage_id=m9 : solo\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertNotContains(t, resp.Text, "[")
	assertNotContains(t, resp.Text, "]")
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "m9")
	assertFormatEqualsDetail(t, resp)
}
```
