---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Stderr mentions mutual exclusion involving session / session-id-from-prompt.

## Exit Code

non-zero

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
	if resp.ExitCode == 0 {
		t.Fatalf("want non-zero exit for --session-id + --session-id-from-prompt; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	msg := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if !strings.Contains(msg, "session") || !strings.Contains(msg, "session-id-from-prompt") {
		// Accept either flag name in the conflict message as long as auto flag is named.
		if !strings.Contains(msg, "session-id-from-prompt") {
			t.Fatalf("error should mention --session-id-from-prompt; got stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
		}
	}
}
```
