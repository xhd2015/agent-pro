## Expected

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
	if !strings.Contains(resp.Stdout, sessions.ForkCommandHelpLine) {
		t.Fatalf("missing ForkCommandHelpLine:\n%s", resp.Stdout)
	}
}
```
