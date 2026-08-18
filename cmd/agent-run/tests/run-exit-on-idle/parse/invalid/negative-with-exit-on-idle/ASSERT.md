## Expected Output

```
---
version: 3
---
Error: --idle-timeout must be a positive duration \(got -1s\)
```

Stderr ends with a trailing newline.

## Expected

- `ParseRunIdle(true, "-1s")` fails.
- Stderr is exactly `Error: --idle-timeout must be a positive duration (got -1s)\n`.
- Exit code is 1.
- `Enabled` is false.

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
		t.Fatal("negative --idle-timeout with --exit-on-idle must be non-zero")
	}
	if resp.ErrString == "" {
		t.Fatal("expected parse error")
	}
	if resp.Enabled {
		t.Fatal("error path must not report enabled=true")
	}
	assert.Output(t, resp.Stderr, `---
version: 3
---
Error: --idle-timeout must be a positive duration \(got -1s\)
`)
}
```
