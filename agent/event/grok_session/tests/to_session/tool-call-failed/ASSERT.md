## Expected
- Wire update line has `status=failed`.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "tool_call_update")
	assertContains(t, resp.Output, `"status":"failed"`)
}
```
