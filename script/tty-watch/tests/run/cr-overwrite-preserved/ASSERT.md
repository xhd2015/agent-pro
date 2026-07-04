## Expected

- PTY capture contains `MARKER_B`.
- Output does **not** concatenate both markers (`MARKER_AMARKER_B`), which indicates
  `\r` was stripped instead of applied as carriage return.

```go
import (
	"strings"
	"testing"
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
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if !strings.Contains(combined, "MARKER_B") {
		t.Fatalf("expected MARKER_B in output, got %q", combined)
	}
	if strings.Contains(combined, "MARKER_AMARKER_B") {
		t.Fatalf("carriage return was stripped (concatenated markers), got %q", combined)
	}
}
```