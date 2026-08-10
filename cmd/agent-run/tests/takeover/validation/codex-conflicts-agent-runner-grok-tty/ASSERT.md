## Expected

- Exit code non-zero.
- `takeover` is recognized (not `unknown command: takeover`).
- Error indicates alias / runner mismatch or conflict between `--codex` and
  `--agent-runner` / `grok-tty` (not a successful takeover path).

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --codex vs --agent-runner=grok-tty, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	combined := resp.Stderr + "\n" + resp.Stdout
	assertTakeoverRecognized(t, combined)
	assertContainsAny(t, combined,
		"conflict",
		"conflicts",
		"mismatch",
		"mutually exclusive",
		"exclusive",
		"cannot use both",
		"incompatible",
		"inconsistent",
	)
	// Anchor to the flags/values under test so unrelated errors stay RED.
	assertContainsAny(t, combined, "codex", "grok-tty", "agent-runner")
	_ = strings.ToLower(combined)
}
```
