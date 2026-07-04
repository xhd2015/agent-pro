## Expected

- Watch capture contains `WATCH_MARKER` from the detached session output.
- Observer receives live PTY bytes (raw stream, not sanitized).

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
	if !strings.Contains(combined, "WATCH_MARKER") {
		t.Fatalf("watch missing WATCH_MARKER, got %q", combined)
	}
}
```