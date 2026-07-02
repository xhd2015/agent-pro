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
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assert.Output(t, resp.Stderr, `
<contains>
<regex>grok-tty:\s*session-\d+</regex>
grok-tty: grok session 550e8400-e29b-41d4-a716-446655440000
grok-tty: grok updates
updates.jsonl
</contains>`)
	assertStderrGrokSession(t, resp.Stderr, stderrGrokUUID, req.GrokUpdatesPath)
}
```