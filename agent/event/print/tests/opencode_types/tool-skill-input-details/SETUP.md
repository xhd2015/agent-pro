# Scenario

**Bug**: opencode `Skill` tool calls print only the `SKILL` header

```
# opencode emits the selected skill and install/use arguments as structured input
Skill tool_use input{name, arguments} -> opencode trace adapter

# compact output keeps those fields visible for trace review
opencode trace adapter -> compact trace printer -> SKILL block with name and arguments
```

## Preconditions
- Native opencode `tool_use` events for the `Skill` tool include useful input
  fields such as the skill name and install arguments.
- The trace printer should show those details below the `SKILL` header.

## Steps
1. Build one native opencode `tool_use` JSONL line for a completed `Skill`
   tool call.
2. Include `name` and `arguments` fields in `part.state.input`.
3. Format the line with `print.FormatTraceLine`.

```go
import (
	"encoding/json"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	line := map[string]any{
		"type":      "tool_use",
		"sessionID": "sess_skill",
		"part": map[string]any{
			"id":     "part_skill",
			"type":   "tool",
			"tool":   "Skill",
			"callID": "call_skill",
			"state": map[string]any{
				"status": "completed",
				"input": map[string]any{
					"name":      "git-fetch",
					"arguments": "skill install --general-agents",
				},
			},
		},
	}
	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("marshal opencode skill event: %v", err)
	}
	req.Line = string(data)
	return nil
}
```
