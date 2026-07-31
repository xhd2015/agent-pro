## Expected

- `RunFocus` is wired (does not fail solely as "unknown command").
- Error, if any, is a resolve/runtime focus error — not an unknown-subcommand parse failure.
- Stderr may contain `Error:` for none/not-found (optional soft).

## Errors

- Allowed: session not found / none found (no AGENT_RUN_HOME inject in this leaf).
- Forbidden: messages indicating command not recognized (`unknown command`).

## Exit Code

- Non-zero on resolve failure is OK for this leaf; GREEN requires implementer wiring + eventual isolated store inject. Classic RED until `RunFocus` exists.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	// Symbol must exist and accept --dry-run without "unknown command".
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "unknown command") {
			t.Fatalf("focus CLI not wired: %v", err)
		}
		// Resolve failure without store is acceptable for dispatch-only leaf.
		t.Logf("RunFocus dry-run resolve error (dispatch OK): %v stderr=%q", err, resp.Stderr)
		return
	}
	// If somehow succeeds without store, still OK for dispatch.
	if strings.Contains(strings.ToLower(resp.Stdout+resp.Stderr), "unknown command") {
		t.Fatalf("unexpected unknown command in output: stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
