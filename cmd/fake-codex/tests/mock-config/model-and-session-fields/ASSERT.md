## Expected
- The hook payload includes configured session and model metadata.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"SessionStart"`)
    assertContains(t, resp.HookLog, `"session_id":"sess_meta"`)
    assertContains(t, resp.HookLog, `"model":"gpt-test"`)
}
```

