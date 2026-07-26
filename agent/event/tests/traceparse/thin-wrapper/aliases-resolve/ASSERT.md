## Expected
- Behavior matches consolidated trace parsing semantics for `thin-wrapper/aliases-resolve`.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	if !resp.AliasOK { t.Fatal("agent_trace.Message should alias traceview.AgentTraceMessage") }

}
```
