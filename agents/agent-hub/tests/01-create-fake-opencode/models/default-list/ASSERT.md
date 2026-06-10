## Expected
- The command succeeds and prints deterministic model names.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, "openai/gpt-5")
}
```

