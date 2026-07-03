## Expected

- `IDs()` includes `grok-tty` and `codex-tty`.
- `stub-tty` absent without `AGENT_RUN_ENABLE_STUB_TTY=1`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	hasGrok, hasCodex := false, false
	for _, id := range resp.ProviderIDs {
		if id == "grok-tty" { hasGrok = true }
		if id == "codex-tty" { hasCodex = true }
	}
	if !hasGrok { t.Fatal("expected grok-tty in IDs()") }
	if !hasCodex { t.Fatal("expected codex-tty in IDs()") }
}
```
