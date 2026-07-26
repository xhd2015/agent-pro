---
label: e2e
---

## Expected

- Browser-side websocket receives initial PTY output.
- Bytes sent by browser are forwarded to upstream PTY.

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
	if !containsAny(resp.WSOutput, req.RegistryTranscript, "echo:"+req.WSInput, "hello from browser") {
		t.Fatalf("websocket did not proxy terminal bytes; output=%q error=%q", resp.WSOutput, resp.WSError)
	}
	if strings.Contains(resp.WSError, "401") {
		t.Fatalf("authorized terminal websocket rejected: %s", resp.WSError)
	}
}
```
