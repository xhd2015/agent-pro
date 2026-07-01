## Expected
- Output JSON array contains one event with `"type":"think"` and text `pondering the question`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"pondering the question"`)
}
```
