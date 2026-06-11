## Expected
- Exit code 0 (yielding questions is not an error).
- Stdout contains the question JSON.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, `"question"`)
    assertContains(t, resp.Stdout, `What is the target port?`)
}
```
