## Expected

- Session created; WS output contains `tty=1`.
- `resp.IsTTY` is true.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected session id")
	}
	if !resp.IsTTY {
		t.Fatalf("expected TTY stdout, got output: %q", resp.WSOutput)
	}
	if !strings.Contains(resp.WSOutput, "tty=1") {
		t.Fatalf("output missing tty=1: %q", resp.WSOutput)
	}
}
```