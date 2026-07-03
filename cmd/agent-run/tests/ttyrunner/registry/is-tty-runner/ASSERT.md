## Expected

- `IsTTYRunner("grok-tty")` and `IsTTYRunner("codex-tty")` are true.
- `IsTTYRunner("fake-codex")` is false.

```go
import (
	"testing"
	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.IsTTYRunner { t.Fatal("expected IsTTYRunner(grok-tty) == true") }
	if ttyrunner.IsTTYRunner("fake-codex") { t.Fatal("expected IsTTYRunner(fake-codex) == false") }
	if !ttyrunner.IsTTYRunner("codex-tty") { t.Fatal("expected IsTTYRunner(codex-tty) == true") }
}
```
