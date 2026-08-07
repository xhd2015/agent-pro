## Expected

```text
Chat history (1 message):
just text
```

- Line is exactly the body (no separators).

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	want := "Chat history (1 message):\njust text\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertNotContains(t, resp.Text, "message_id=")
	assertNotContains(t, resp.Text, "[")
	assertFormatEqualsDetail(t, resp)
}
```
