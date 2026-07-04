## Expected

- PTY capture contains `TTY_WATCH_INTERRUPTED` after harness sends `\x03`.
- Child trap handler ran (interrupt reached PTY child, not just client detach).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if !strings.Contains(combined, "TTY_WATCH_INTERRUPTED") {
		t.Fatalf("Ctrl-C did not reach child, output %q", combined)
	}
}
```