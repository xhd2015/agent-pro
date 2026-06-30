## Expected

- REST-created session runs `echo hello`.
- WS output contains `hello`.

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
	if !strings.Contains(resp.WSOutput, "hello") {
		t.Fatalf("expected hello in output, got: %q", resp.WSOutput)
	}
}
```