## Expected

- Raw PTY output is exactly `yes\r\n[Terminal exited]\r\n` (LF-only newlines smear column on interactive TTY).
- No bare `\n` without a preceding `\r` between content lines.
- Host shell prompt after exit must start at column 0 (requires trailing `\r\n`).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.TimedOut {
		t.Fatalf("timed out, output %q", resp.Combined)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d output %q", resp.ExitCode, resp.Combined)
	}
	raw := resp.Stdout
	if raw == "" {
		raw = resp.Combined
	}
	want := "yes\r\n[Terminal exited]\r\n"
	if raw != want {
		t.Fatalf("raw PTY output = %q, want %q", raw, want)
	}
	if ttywatchtest.HasBareLFNewlines(raw) {
		t.Fatalf("output has bare LF newlines (column drift on raw TTY), got %q", raw)
	}
}
```