---
label: e2e
---

## Expected Output

Stderr includes internal ptywrap id plus discovered grok session diagnostics:

```
<contains>
grok-tty: session-
grok-tty: grok session 550e8400-e29b-41d4-a716-446655440000
grok-tty: grok updates
updates.jsonl
</contains>
```

## Expected

- Exit code 0.
- Stderr matches `grok-tty: session-\d+` (internal id).
- Stderr contains `grok-tty: grok session 550e8400-e29b-41d4-a716-446655440000`.
- Stderr contains `grok-tty: grok updates` with absolute path ending in
  `updates.jsonl` under the seeded grok session dir.

## Exit Code

0

```go
import (
	"regexp"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	re := regexp.MustCompile(`(?m)^grok-tty:\s*session-\d+\s*$`)
	if !re.MatchString(resp.Stderr) {
		t.Fatalf("stderr missing grok-tty: session-N; stderr:\n%s", resp.Stderr)
	}
	assertStderrGrokSession(t, resp.Stderr, stderrGrokUUID, req.GrokUpdatesPath)
}
```
