## Expected

- PTY capture contains `MARKER_B`.
- Output does **not** concatenate both markers (`MARKER_AMARKER_B`), which indicates
  `\r` was stripped instead of applied as carriage return.
- Wire format has no bare `\n` without a preceding `\r` (CRLF normalization on raw TTY).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.TimedOut {
		t.Fatalf("cr overwrite probe hung, output %q", resp.Combined)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d output %q", resp.ExitCode, resp.Combined)
	}
	raw := resp.Stdout
	if raw == "" {
		raw = resp.Combined
	}
	if !strings.Contains(raw, "MARKER_B") {
		t.Fatalf("expected MARKER_B in output, got %q", raw)
	}
	if strings.Contains(raw, "MARKER_AMARKER_B") {
		t.Fatalf("carriage return was stripped (concatenated markers), got %q", raw)
	}
	if ttywatchtest.HasBareLFNewlines(raw) {
		t.Fatalf("output has bare LF newlines (column drift on raw TTY), got %q", raw)
	}
	if !strings.HasSuffix(raw, "[Terminal exited]\r\n") {
		t.Fatalf("exit marker must end with CRLF on raw TTY, got %q", raw)
	}
}
```