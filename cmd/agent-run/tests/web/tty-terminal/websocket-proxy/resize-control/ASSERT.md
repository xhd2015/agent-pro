---
label: e2e
---

## Expected

- Resize message is forwarded to the upstream terminal websocket.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.WSOutput, "resize-ok") && !strings.Contains(resp.WSResize, `"cols":100`) {
		t.Fatalf("resize was not forwarded; output=%q resize=%q error=%q", resp.WSOutput, resp.WSResize, resp.WSError)
	}
}
```
