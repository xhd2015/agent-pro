## Expected

- Foreground dir is the `--dir` override, not session cwd.
- Still no new window.

## Side Effects

- One recorded foreground.

## Errors

- None.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoOpen(t, resp)
	call := assertOneForeground(t, resp)
	if call.Dir != req.OverrideDir {
		t.Fatalf("foreground dir: got %q, want override %q", call.Dir, req.OverrideDir)
	}
}
```
