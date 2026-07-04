## Expected

- PTY output shows `yes` then `[Terminal exited]`, both starting at column 0.
- No leading blank line before `yes` (relay must not prepend `\n` before snapshot text).
- No leading spaces before `[Terminal exited]` (common `\r`/LF layout bug).
- Output ends with `\n` so the host shell prompt does not render far to the right.

```go
import (
	"reflect"
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
	if err := ttywatchtest.NoLeadingBlankLine(raw); err != nil {
		t.Fatalf("output must not start with blank line before yes: %v; raw %q", err, raw)
	}
	if ttywatchtest.HasIndentedExitMarker(raw) {
		t.Fatalf("exit marker not left-aligned, got %q", raw)
	}
	lines, err := ttywatchtest.ContentLinesAtColumnZero(raw)
	if err != nil {
		t.Fatalf("output not left-aligned: %v; raw %q", err, raw)
	}
	want := []string{"yes", "[Terminal exited]"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("content lines = %v, want %v; raw %q", lines, want, raw)
	}
	if !ttywatchtest.EndsWithNewline(raw) {
		t.Fatalf("output must end with newline so host prompt starts at column 0, got %q", raw)
	}
}
```