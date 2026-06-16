## Preconditions
- The `print.FormatAgentEvent` function formats `AgentEvent` structs into
  human-readable strings. Each leaf constructs an `AgentEvent` with specific
  fields and verifies the formatted output contains expected markers.

## Steps
1. Build an `AgentEvent` from `req.Type`, `req.Text`, `req.Tool`,
   `req.Output`, `req.ExitCode`, and `req.Changes`.
2. Call `print.FormatAgentEvent(event)`.
3. Return the formatted string in `resp.Output`.

## Context
- `Req.Type` — the `ActionType` (tool_call, message, think, error,
  step_start, step_finish, done, sleep, or any string for unknown types).
- `Req.Text` — optional text associated with the event.
- `Req.Tool` — tool name (for tool_call events).
- `Req.Output` — tool output (for tool_call events).
- `Req.ExitCode` — pointer to int; nil means no exit code.
- `Req.Changes` — slice of `FileChange` (path + kind).

```go
import (
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"

	"github.com/xhd2015/agent-pro/agent/event/print"
)

type Request struct {
	Type     types.ActionType
	Text     string
	Tool     string
	Output   string
	ExitCode *int
	Changes  []types.FileChange
}

type Response struct {
	Output string
}

func Setup(t *testing.T, req *Request) error {
	_ = assertContains
	_ = assertNotContains
	_ = assertEmpty
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	event := types.AgentEvent{
		Type:     req.Type,
		Text:     req.Text,
		Tool:     req.Tool,
		Output:   req.Output,
		ExitCode: req.ExitCode,
		Changes:  req.Changes,
	}
	return &Response{Output: print.FormatAgentEvent(event)}, nil
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got string, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}

func assertEmpty(t *testing.T, got string) {
	t.Helper()
	if strings.TrimSpace(got) != "" {
		t.Fatalf("expected empty output, got:\n%s", got)
	}
}
```
