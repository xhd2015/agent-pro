## Expected

- Exit code non-zero.
- Error mentions 10s timeout and provider reason.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode == 0 { t.Fatal("expected non-zero exit on send timeout") }
	combined := resp.MultiAttachProbe.SendTimeoutReason
	if !strings.Contains(combined, "timed out") && !strings.Contains(combined, "10s") {
		t.Fatalf("expected timeout message, got: %s", combined)
	}
}
```
