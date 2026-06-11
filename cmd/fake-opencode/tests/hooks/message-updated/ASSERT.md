## Expected
- The `message.updated` hook fires with prompt text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"message.updated"`)
    assertContains(t, resp.HookLog, `"text":"prompt text"`)
}
```

