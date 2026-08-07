## Expected

- `BodiesTruncated=1`
- `Shown=3`, `SourceCount=3`
- Truncated middle body is `abc…` (3 letters + marker)
- Short and exact bodies unchanged (`ab`, `xyzw`)

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	assertEqualInt(t, "BodiesTruncated", resp.Detail.BodiesTruncated, 1)
	assertEqualInt(t, "Shown", resp.Detail.Shown, 3)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 3)
	assertContains(t, resp.Text, "ab")
	assertContains(t, resp.Text, "abc"+truncationMarker)
	assertContains(t, resp.Text, "xyzw")
	assertNotContains(t, resp.Text, "abcdefgh")
	assertFormatEqualsDetail(t, resp)
}
```
