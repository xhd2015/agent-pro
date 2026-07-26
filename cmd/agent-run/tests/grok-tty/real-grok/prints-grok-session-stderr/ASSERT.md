---
label: e2e, grok
explanation: Requires real grok CLI on PATH; verifies stderr grok session diagnostics and live stdout.
---

## Expected

- Exit code 0.
- Stderr contains `grok-tty: grok session` with a UUID-like token.
- Stderr contains `grok-tty: grok updates` with path ending in `updates.jsonl`.
- Stdout is **non-empty** before the 60s stream-probe timeout (live streaming).

## Exit Code

0

```go
import (
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

var grokSessionStderrRE = regexp.MustCompile(`grok-tty:\s*grok session\s+[0-9a-fA-F-]{8,}`)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(resp.Stderr), "banner not detected") {
		t.Fatalf("grok banner not detected:\n%s", resp.Stderr)
	}
	assertSuccess(t, resp)

	assert.Output(t, resp.Stderr, `
<contains>
grok-tty: grok session
grok-tty: grok updates
updates.jsonl
</contains>`)
	if !grokSessionStderrRE.MatchString(resp.Stderr) {
		t.Fatalf("stderr missing grok session uuid line; stderr:\n%s", resp.Stderr)
	}

	stdout := strings.TrimSpace(resp.Stdout)
	if stdout == "" {
		t.Fatalf("expected non-empty stdout from live streaming; stderr:\n%s", resp.Stderr)
	}
	if !resp.StreamProbeSeen {
		t.Fatalf("expected stdout content before 60s timeout (live stream); stdout:\n%s", resp.Stdout)
	}
}
```