## Expected

- Non-zero exit before session starts.
- Stderr mentions invalid session id and the rejected value.

## Errors

- `run: invalid session id ".bad"`

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
	if !strings.Contains(resp.Combined, `invalid session id`) {
		t.Fatalf("expected invalid session id error, got combined %q", resp.Combined)
	}
	if !strings.Contains(resp.Combined, ".bad") {
		t.Fatalf("expected rejected id in error, got combined %q", resp.Combined)
	}
}
```