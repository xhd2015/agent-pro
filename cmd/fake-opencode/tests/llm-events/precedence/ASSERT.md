## Expected
- The command succeeds.
- stdout contains the text from `llm_events`.
- stdout does NOT contain the text from `stdout_events`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"from llm_events"`)
    // stdout_events content must not appear — llm_events wins
    if strings.Contains(resp.Stdout, `"from stdout_events"`) {
        t.Fatalf("stdout_events content leaked despite llm_events precedence:\n%s", resp.Stdout)
    }
}
```
