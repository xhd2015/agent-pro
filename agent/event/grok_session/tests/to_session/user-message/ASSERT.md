## Expected
- Wire contains `user_message_chunk` with text `run ls`.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "user_message_chunk")
	assertContains(t, resp.Output, "run ls")
}
```
