---
label: e2e
---

## Expected

- Help text mentions `--session-id`.
- Stdout ends with trailing `\n` (CLI user-facing lines).

## Exit Code

0 (or help exit policy of the binary — treat success if `--session-id` is documented)

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
	out := resp.Stdout + resp.Stderr
	if !strings.Contains(out, "--session-id") {
		t.Fatalf("run --help should document --session-id; got:\n%s", out)
	}
	// Prefer stdout ends with newline when help is on stdout.
	if resp.Stdout != "" && !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("stdout should end with \\n; got %q", resp.Stdout)
	}
}
```
