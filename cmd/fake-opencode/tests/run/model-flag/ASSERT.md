## Expected
- The hook payload includes the model flag.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"model":"openai/gpt-5"`)
}
```

