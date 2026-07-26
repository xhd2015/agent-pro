## Expected

- `grok-tty` → `grok-tty-registry`.
- `codex-tty` → `codex-tty-registry`.

```go
import (
	"testing"
	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.RegistryDir != "grok-tty-registry" {
		t.Fatalf("grok-tty RegistryDir: got %q want grok-tty-registry", resp.RegistryDir)
	}
	p, ok := ttyrunner.Get("codex-tty")
	if !ok { t.Fatal("codex-tty not registered") }
	if p.RegistryDir != "codex-tty-registry" {
		t.Fatalf("codex-tty RegistryDir: got %q want codex-tty-registry", p.RegistryDir)
	}
}
```
