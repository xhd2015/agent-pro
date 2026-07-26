---
label: unit
explanation: no search hits -> {"channels":[]} with exit 0
---

## Expected Output

```json
{"channels":[]}
```

## Expected

- Exit code 0 (not an error).
- JSON empty channels array with trailing newline.
- Stderr empty.

## Exit Code

0

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	var doc struct {
		Channels []any `json:"channels"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	if len(doc.Channels) != 0 {
		t.Fatalf("channels len = %d, want 0: %+v", len(doc.Channels), doc.Channels)
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
}
```
