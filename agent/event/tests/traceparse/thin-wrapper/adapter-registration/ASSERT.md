## Expected
- Behavior matches consolidated trace parsing semantics for `thin-wrapper/adapter-registration`.

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

	if !resp.OK { t.Fatal("expected opencode adapter registered via agent_trace import") }
	assertContains(t, resp.Output, "registered")

}
```
