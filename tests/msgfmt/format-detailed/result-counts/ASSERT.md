## Expected

- `SourceCount=5`, `Shown=2`, `OldestDropped=3`
- Header `Chat history (showing 2 of 5):`
- Bodies `t4` and `t5` present; `t1` absent

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 5)
	assertEqualInt(t, "Shown", resp.Detail.Shown, 2)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 3)
	assertContains(t, resp.Text, "Chat history (showing 2 of 5):")
	assertContains(t, resp.Text, "t4")
	assertContains(t, resp.Text, "t5")
	assertNotContains(t, resp.Text, "t1")
	assertFormatEqualsDetail(t, resp)
}
```
