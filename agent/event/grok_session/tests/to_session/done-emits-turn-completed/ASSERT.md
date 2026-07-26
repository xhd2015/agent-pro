## Expected
- Wire contains `turn_completed`.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWireHasSessionUpdate(t, resp.WireLines, "turn_completed")
}
```
