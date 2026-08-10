## Expected

- Exit code non-zero.
- `takeover` is recognized (not `unknown command: takeover`).
- Combined output indicates mutual exclusion between `--grok` and `--codex`
  (or equivalent "cannot use both" / conflict wording).
- Error anchors to at least one of the two flags so unrelated failures stay RED.

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
		t.Fatalf("expected non-zero exit for --grok/--codex mutex, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	combined := resp.Stderr + "\n" + resp.Stdout
	assertTakeoverRecognized(t, combined)
	assertContainsAny(t, combined,
		"mutually exclusive",
		"exclusive",
		"cannot use both",
		"cannot be used with",
		"not both",
		"conflict",
		"incompatible",
	)
	// Anchor to the provider alias flags under test.
	assertContainsAny(t, combined, "grok", "codex")
	_ = strings.ToLower(combined)
}
```
