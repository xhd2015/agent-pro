## Expected

- Parent help line names `resolve` and contextual session id.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	if !strings.Contains(resp.Stdout, sessions.ResolveCommandHelpLine) {
		t.Fatalf("stdout missing ResolveCommandHelpLine:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "resolve") {
		t.Fatalf("stdout missing resolve:\n%s", resp.Stdout)
	}
}
```
