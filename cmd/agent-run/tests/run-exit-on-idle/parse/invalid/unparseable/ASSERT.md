## Expected Output

```
---
version: 3
---
Error: invalid value for --idle-timeout: nope
```

Stderr ends with a trailing newline.

## Expected

- `ParseRunIdle(false, "nope")` fails before normalize.
- Stderr is exactly `Error: invalid value for --idle-timeout: nope\n`.
- Exit code is 1.

## Side Effects

- None (no TTY).

## Errors

- Parse error required.

## Exit Code

1

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatal("unparseable --idle-timeout must be non-zero")
	}
	if resp.ErrString == "" {
		t.Fatal("expected parse error")
	}
	assert.Output(t, resp.Stderr, `---
version: 3
---
Error: invalid value for --idle-timeout: nope
`)
}
```
