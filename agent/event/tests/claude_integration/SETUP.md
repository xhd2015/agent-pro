# Scenario

**Feature**: claude headless integration runs the real claude binary and parses its stream-json output

```
# run claude -p <prompt> --output-format stream-json --verbose, scan stdout lines
Run{Target:"claude_headless", Prompt} -> runClaudeHeadless -> claude -p ... -> []StreamEvent -> FromClaude -> resp.Output

# skip when CLAUDE_SKIP_INTEGRATION=1 or claude not on PATH; error when explicit ClaudePath missing
```

## Preconditions
- A `claude` binary is available on `PATH` (or `ClaudePath` is configured) for the live leaf.
- The root `runClaudeHeadless` helper handles binary lookup, skip, and error cases.

## Steps
1. Set `req.Target = "claude_headless"` so the root `Run` dispatches to `runClaudeHeadless`.
2. Leaf SETUPs populate `req.Prompt` and (for the not-found leaf) `req.ClaudePath`.

## Context
- The root `runClaudeHeadless` (defined in `../SETUP.md`) runs `claude -p <Prompt> --output-format stream-json --verbose`, scans stdout into `[]claude_types.StreamEvent`, calls `FromClaude`, and marshals to `Output`.
- If `claude` is not on PATH and `ClaudePath` is empty, the test is skipped (`t.Skip`).
- If `ClaudePath` is explicitly set and `os.Stat` fails, a non-nil error is returned.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "claude_headless"
	if req.SessionID == "" {
		req.SessionID = "sess_claude"
	}
	return nil
}
```
