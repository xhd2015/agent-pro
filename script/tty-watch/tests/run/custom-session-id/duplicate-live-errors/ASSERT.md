## Expected

- Second `run --session-id` exits non-zero.
- Stderr (or combined) mentions session id already in use.

## Errors

- `run: session id "test-with-grok" already in use`

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertNonZeroExit(t, resp)
	lower := strings.ToLower(resp.Combined)
	if !strings.Contains(lower, "already in use") {
		t.Fatalf("expected already in use error, got combined %q", resp.Combined)
	}
	if !strings.Contains(resp.Combined, "test-with-grok") {
		t.Fatalf("expected session id in error, got combined %q", resp.Combined)
	}
}
```