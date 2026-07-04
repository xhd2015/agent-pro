## Expected

- TTY watch must not emit a plain-text screen snapshot followed by raw true-color SGR
  incremental updates. That mixed protocol leaves the observer terminal out of sync
  and shows `[38;2;...` garbage in the input area.
- When incremental `\x1b[38;2;` updates are present, the initial screen replay must
  also pass through as raw terminal protocol (`\x1b[?25l` or `\x1b[?1049h`), not
  converted plain-text box drawing.
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
	hasPlainBox := strings.Contains(combined, "╭")
	hasRawSGR := strings.Contains(combined, "\x1b[38;2;")
	if !hasPlainBox || !hasRawSGR {
		t.Fatalf("fixture did not produce mixed plain+SGR output to test, got %q", combined)
	}
	plainIdx := strings.Index(combined, "╭")
	sgrIdx := strings.Index(combined, "\x1b[38;2;")
	if plainIdx >= 0 && sgrIdx >= 0 && plainIdx < sgrIdx {
		t.Fatalf("watch emitted plain-text snapshot before raw SGR updates (input-box [38;2; garbage), got %q", combined)
	}
}
```