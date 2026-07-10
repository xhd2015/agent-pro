---
label: unit
explanation: "auth.test invalid_auth -> auth failed: exit 1"
---

## Expected

- Exit code 1.
- Stderr contains `auth failed:`.
- Stdout must not contain full raw token.
- No successful `ok: true` status block required on stdout (may be empty or partial).

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStderrContains(t, resp, "auth failed:")
	if strings.Contains(resp.Stdout, slackTestToken) {
		t.Fatalf("stdout must not contain raw bot token %q:\n%s", slackTestToken, resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "ok: true") {
		t.Fatalf("unexpected ok: true on failed auth:\n%s", resp.Stdout)
	}
}
```
