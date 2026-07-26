---
label: e2e
---

## Expected
- The event's `aggregated_output` contains `"hello real codex"` (real bash output).
- The `exit_code` is 0.

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    stdout := resp.Stdout
    if !strings.Contains(stdout, "hello real codex") {
        t.Fatalf("expected real bash output 'hello real codex' in stdout, got:\n%s", stdout)
    }
    if !strings.Contains(stdout, `"aggregated_output"`) {
        t.Fatalf("expected aggregated_output field, got:\n%s", stdout)
    }
    // Non-zero exit would mean bash failed
    if strings.Contains(stdout, `"exit_code":1`) || strings.Contains(stdout, `"exit_code":-1`) {
        t.Fatalf("expected exit_code 0, got:\n%s", stdout)
    }
}
```
