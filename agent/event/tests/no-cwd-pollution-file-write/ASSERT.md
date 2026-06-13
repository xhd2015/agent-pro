## Expected
- No files are created in the working directory.
- Only "DONE" is printed (no "FILE:" lines).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "DONE")
	assertNotContains(t, resp.Output, "FILE:")
}
```
