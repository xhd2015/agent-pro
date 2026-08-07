## Expected

- `LastMessageID="m-new"`
- Only `new` body shown; header `Chat history (showing 1 of 3):`

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "m-new")
	assertEqualInt(t, "Shown", resp.Detail.Shown, 1)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 3)
	assertContains(t, resp.Text, "Chat history (showing 1 of 3):")
	assertContains(t, resp.Text, "new")
	assertNotContains(t, resp.Text, "old")
	assertFormatEqualsDetail(t, resp)
}
```
