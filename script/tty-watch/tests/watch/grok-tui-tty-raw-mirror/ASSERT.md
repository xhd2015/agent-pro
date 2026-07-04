## Expected

- Watch PTY capture contains raw terminal escape sequences (`\x1b[`) so the observer
  terminal can mirror the grok alternate-screen layout.
- Watch must contain the alternate-screen entry sequence `\x1b[?1049h` (or equivalent
  full-screen TUI protocol), not only plain-text box-drawing characters stripped of escapes.
- Observer mode is readonly.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if !strings.Contains(combined, "\x1b[") {
		t.Fatalf("watch TTY did not receive raw terminal protocol (expected mirror like main session), got %q", combined)
	}
	if !strings.Contains(combined, "\x1b[?1049h") {
		t.Fatalf("watch TTY missing alternate-screen entry sequence, got %q", combined)
	}
}
```