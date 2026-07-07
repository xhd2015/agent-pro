## Expected

- Websocket receives initial PTY transcript and resize acknowledgement.
- Upstream ptywrap connection uses `attach_mode=attach` (not snapshot).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.WSError != "" {
		t.Fatalf("websocket error: %s", resp.WSError)
	}
	if !containsAny(resp.WSOutput, req.RegistryTranscript, "resize-ok") {
		t.Fatalf("websocket did not relay attach output; got=%q", resp.WSOutput)
	}
	if req.RegistryResize == "" || !strings.Contains(req.RegistryResize, "resize") {
		t.Fatalf("resize control did not reach upstream PTY; resize=%q", req.RegistryResize)
	}
}
```
