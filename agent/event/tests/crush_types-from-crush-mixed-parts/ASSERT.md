## Expected
- Two canonical AgentEvents:
  - Type `message` with text "let me run that command"
  - Type `tool_call` with tool `bash` and command `ls`

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"let me run that command"`)
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
	assertContains(t, resp.Output, `"ls"`)
}
```
