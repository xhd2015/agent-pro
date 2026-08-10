## Expected

- Exit code non-zero.
- `takeover` is recognized (not `unknown command: takeover` alone).
- Error indicates an unknown / unrecognized / invalid flag (or flag parse failure).

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
		t.Fatalf("expected non-zero exit for unknown flag, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	combined := resp.Stderr + "\n" + resp.Stdout
	assertTakeoverRecognized(t, combined)
	assertContainsAny(t, combined,
		"unknown",
		"unrecognized",
		"invalid",
		"flag",
		"not-a-real-takeover-flag",
	)
	_ = strings.ToLower(combined)
}
```
