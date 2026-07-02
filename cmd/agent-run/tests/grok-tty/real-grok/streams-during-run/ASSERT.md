---
label: grok
explanation: Requires real grok CLI on PATH; verifies live streaming during run ls.
---

## Expected

- Exit code 0.
- Stdout is **non-empty before** the 60s stream-probe timeout (live streaming, not
  silent-until-exit).
- Output contains `agent` (directory entry) or listing tokens (`drwx`, `total`).

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
	if strings.Contains(strings.ToLower(resp.Stderr), "banner not detected") {
		t.Fatalf("grok banner not detected:\n%s", resp.Stderr)
	}
	assertSuccess(t, resp)

	stdout := strings.TrimSpace(resp.Stdout)
	if stdout == "" {
		t.Fatalf("expected non-empty stdout from live streaming; stderr:\n%s", resp.Stderr)
	}
	if !resp.StreamProbeSeen {
		t.Fatalf("expected stdout content before 60s timeout (live stream); stdout:\n%s", resp.Stdout)
	}

	lower := strings.ToLower(resp.Stdout)
	hasListing := strings.Contains(lower, "agent") ||
		strings.Contains(lower, "drwx") ||
		strings.Contains(lower, "total ")
	if !hasListing {
		t.Fatalf("expected listing or agent token in streamed stdout:\n%s", resp.Stdout)
	}
}
```