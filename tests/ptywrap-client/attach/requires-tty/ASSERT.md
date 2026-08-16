## Expected

- `Attach` returns non-nil error.
- Error message mentions interactive terminal or TTY requirement.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.AttachErr == "" {
		t.Fatal("expected attach error for non-TTY stdin")
	}
	lower := strings.ToLower(resp.AttachErr)
	if !strings.Contains(lower, "tty") && !strings.Contains(lower, "terminal") && !strings.Contains(lower, "interactive") {
		t.Fatalf("attach error should mention tty/terminal, got %q", resp.AttachErr)
	}
}
```