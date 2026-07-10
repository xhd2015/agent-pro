---
label: unit
explanation: "search empty after soft-skip private; warning with see: topic + empty JSON exit 0"
---

## Expected Output

```json
{"channels":[]}
```

## Expected

- Exit code 0 (no match is not an error; soft-skip is not a hard failure).
- JSON empty channels array with trailing newline.
- Stderr contains
  `warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope`.

## Exit Code

0

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStderrContains(t, resp, "warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope")
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
