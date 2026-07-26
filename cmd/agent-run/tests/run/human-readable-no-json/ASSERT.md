---
label: e2e
---

## Expected

- Exit code 0.
- Stdout is human-readable via `agent/event/print` — not exclusively JSON lines.
- At least one stdout line is not a JSON object line.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if strings.TrimSpace(resp.Stdout) == "" {
		t.Fatal("expected non-empty human-readable stdout")
	}
	allJSON := true
	for _, line := range strings.Split(resp.Stdout, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !isJSONObjectLine(line) {
			allJSON = false
			break
		}
	}
	if allJSON {
		t.Fatalf("expected human-readable output (not all JSON lines), got:\n%s", resp.Stdout)
	}
}
```