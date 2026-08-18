## Expected

- `RunSessions` returns nil.
- Stdout is the locked resolve help (trailing `\n`), documenting Usage,
  `--grok-session-id`, and `--json`.

## Expected Output

```
Usage: agent-run sessions resolve --grok-session-id ID
       agent-run sessions resolve [--json] --grok-session-id ID

Resolve an agent-run session id from a Grok provider runner_session_id.
Read-only; never creates a session.

Options:
  --grok-session-id ID   provider UUID (meta.runner grok|grok-tty)
  --json                 print session meta as JSON
  -h, --help             show help
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil {
		t.Fatalf("resolve --help error: %v", resp.Err)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
Usage: agent-run sessions resolve --grok-session-id ID
       agent-run sessions resolve \[--json\] --grok-session-id ID

Resolve an agent-run session id from a Grok provider runner_session_id\.
Read-only; never creates a session\.

Options:
  --grok-session-id ID   provider UUID \(meta\.runner grok\|grok-tty\)
  --json                 print session meta as JSON
  -h, --help             show help
`)
}
```
