## Expected
- Output contains `"type":"turn_start"`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"turn_start"`)
}
```
