# Scenario

**Feature**: a system init line maps to one ActionStepStart

```
# system subtype=init signals run begin
{"type":"system","subtype":"init",...} -> FromClaude -> ActionStepStart
```

## Preconditions
- `FromClaude` emits exactly one `ActionStepStart` for a `system` `init` event.

## Steps
1. Provide a single `system` init NDJSON line.
2. Call `FromClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeInput = `{"type":"system","subtype":"init","cwd":"/tmp","session_id":"sess_claude","model":"claude-sonnet","tools":[],"permissionMode":"default"}`
	return nil
}
```
