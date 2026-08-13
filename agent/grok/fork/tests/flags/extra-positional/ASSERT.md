## Expected

- Error mentions unexpected/extra argument (or the leftover token).
- Exit 1; no launch.

## Side Effects

- None.

## Errors

- unexpected / extra argument

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	if resp == nil || resp.Err == nil {
		t.Fatal("expected error for extra positional")
	}
	msg := strings.ToLower(resp.Err.Error())
	if !strings.Contains(msg, "unexpected") && !strings.Contains(msg, "extra") &&
		!strings.Contains(msg, "argument") && !strings.Contains(resp.Err.Error(), "leftover") {
		t.Fatalf("error %q should mention unexpected/extra argument", resp.Err)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("ExitCode=%d, want 1", resp.ExitCode)
	}
	assertNoOpen(t, resp)
}
```
