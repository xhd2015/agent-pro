## Expected

- Watch capture contains the grok-like prompt text `Grok Build ›`.
- Watch capture must **not** contain CSI, OSC, or C0 control sequences (the stray
  characters that appear in the input box on a real terminal).
- Observer mode is readonly and visually mirrors the running session.

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
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if !strings.Contains(combined, "Grok Build") {
		t.Fatalf("watch missing grok-like prompt, got %q", combined)
	}
	if ttywatchtest.ContainsANSIEscape(combined) {
		t.Fatalf("watch leaked control/escape sequences into observer output (input-box garbage), got %q", combined)
	}
}
```