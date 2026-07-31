## Expected

- `FocusSession` returns nil error.
- Chosen candidate TTY matches `/dev/ttys148` (or normalized equal).
- Chosen `Ref.WindowID` is `win-1`, `TabIndex` 2.
- `FocusITerm` called exactly once with that ref (WindowID `win-1`).

## Errors

- None.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	if iterm2.NormalizeTTY(resp.Chosen.TTY) != iterm2.NormalizeTTY("/dev/ttys148") &&
		iterm2.NormalizeTTY(resp.Chosen.Ref.TTY) != iterm2.NormalizeTTY("/dev/ttys148") {
		t.Fatalf("chosen TTY = %+v, want ttys148", resp.Chosen)
	}
	if resp.Chosen.Ref.WindowID != "win-1" || resp.Chosen.Ref.TabIndex != 2 {
		t.Fatalf("chosen ref = %+v, want win-1 tab 2", resp.Chosen.Ref)
	}
	if len(resp.FocusCalls) != 1 {
		t.Fatalf("FocusITerm calls = %d, want 1; calls=%+v", len(resp.FocusCalls), resp.FocusCalls)
	}
	if resp.FocusCalls[0].WindowID != "win-1" {
		t.Fatalf("FocusITerm ref = %+v, want win-1", resp.FocusCalls[0])
	}
}
```
