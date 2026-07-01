## Expected
- All four event-type constant values are correct (`system`, `assistant`, `user`, `result`).
- The marshaled `StreamEvent` JSON carries `"type":"assistant"`, `"session_id":"sess_claude"`, and a `message.content` array with a `text` block whose text is `pong`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "EventSystem=system")
	assertContains(t, resp.Output, "EventAssistant=assistant")
	assertContains(t, resp.Output, "EventUser=user")
	assertContains(t, resp.Output, "EventResult=result")

	assertContains(t, resp.Output, `"type":"assistant"`)
	assertContains(t, resp.Output, `"session_id":"sess_claude"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"text":"pong"`)
}
```
