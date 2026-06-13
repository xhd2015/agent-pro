## Expected
- No files are created in the working directory.
- Only "DONE" is printed (no "FILE:" lines).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "DONE")
	assertNotContains(t, resp.Stdout, "FILE:")
}
```
