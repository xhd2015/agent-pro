## Expected
- The command succeeds.
- stdout contains reasoning events (started+completed) with "t1".
- stdout contains a message completed event with "hello codex".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"type":"message"`)
    assertContains(t, resp.Stdout, `"t1"`)
    assertContains(t, resp.Stdout, `"hello codex"`)
}
```
