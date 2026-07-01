## Expected
- Output JSON array contains one event with `"type":"step_start"`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_start"`)
}
```
