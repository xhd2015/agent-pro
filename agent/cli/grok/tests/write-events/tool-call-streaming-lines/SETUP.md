# Scenario

**Feature**: grok tool_started/tool_completed lines convert to tool_call AgentEvents

## Preconditions
- Grok sessions emit `tool_started` / `tool_completed` lines (see `~/.grok/sessions/.../events.jsonl`).
- `GrokEventWriter` should convert those into `ActionToolCall` AgentEvent lines for `events.jsonl`.
- Today only `thought`, `text`, and `end` streaming lines are converted; tool activity is dropped.

## Steps
1. Set `req.GrokLines` to a mixed grok stream: thought, Read tool, Grep tool, assistant text.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokLines = []string{
		`{"type":"thought","data":"I'll read the requirement and search the tree."}`,
		`{"type":"tool_started","tool_name":"Read"}`,
		`{"type":"tool_completed","tool_name":"Read","duration_ms":1,"outcome":"success"}`,
		`{"type":"tool_started","tool_name":"Grep"}`,
		`{"type":"tool_completed","tool_name":"Grep","duration_ms":2,"outcome":"success"}`,
		`{"type":"text","data":"I'll implement the DOCTEST.md version and layout changes."}`,
	}
	return nil
}
```