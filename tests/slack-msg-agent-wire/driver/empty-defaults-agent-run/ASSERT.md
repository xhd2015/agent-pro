## Expected

- Source still exposes `agentRunBinary` (or equivalent) for binary resolution.
- Empty-driver fallback to **`agent-run`** is explicit in source:
  - comment/docs mentioning empty → agent-run, and/or
  - literal `"agent-run"`, and/or
  - `DriverBinary` field filled from `agentRunBinary()` with empty meaning library default.
- Prefer: contains `agent-run` (compat name) and does not hardcode a different default binary as the only empty fallback.

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	src := resp.AgentSrc
	if !resp.HasAgentRunBinary && !strings.Contains(src, "DriverBinary") {
		t.Fatal("agent.go must resolve binary via agentRunBinary and/or DriverBinary")
	}
	// Empty env must default to agent-run (compat). Require the token in source
	// (comment, string default, or library-default documentation).
	if !strings.Contains(src, "agent-run") {
		t.Fatal(`agent.go must document or default empty driver to "agent-run"`)
	}
	// Must not claim empty defaults to agentrunbridge only without agent-run name
	// after cutover; interactive path should not be the sole bridge open.
	_ = resp.HasDriverBinary
}
```
