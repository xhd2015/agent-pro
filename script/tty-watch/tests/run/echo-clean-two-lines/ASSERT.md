## Expected

- Visible content is exactly two lines: `yes` then `[Terminal exited]`.
- No leading blank line before `yes` in raw PTY output.
- No alternate-screen exit prefix (`\x1b[?1049l`) in PTY output — that prefix
  smears a cleared terminal with blank lines before short-command output.

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
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	raw := resp.Stdout
	if raw == "" {
		raw = combined
	}
	if err := ttywatchtest.NoLeadingBlankLine(raw); err != nil {
		t.Fatalf("output must not start with blank line before yes: %v; raw %q", err, raw)
	}
	if ttywatchtest.HasAlternateScreenExitPrefix(combined) {
		t.Fatalf("alternate-screen exit prefix smears terminal, got %q", combined)
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
}
```