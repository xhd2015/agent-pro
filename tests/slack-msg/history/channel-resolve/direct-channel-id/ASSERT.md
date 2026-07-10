---
label: unit
explanation: history with direct channel ID
---

## Expected

- Exit code 0.
- At least one human history line (oldest first).
- Stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "[1710000001.000100] U_OLDER: first message") {
		t.Fatalf("stdout missing oldest message:\n%s", resp.Stdout)
	}
	if !strings.HasPrefix(strings.TrimSpace(resp.Stdout), "[1710000001.000100]") {
		t.Fatalf("first line should be oldest message:\n%s", resp.Stdout)
	}
}
```
