# Scenario

**Bug**: opencode `TodoWrite` tool calls print only a todo/plan header

```
# opencode emits todos as structured input items
TodoWrite tool_use input{todos[]} -> opencode trace adapter

# compact output keeps item content and status visible for trace review
opencode trace adapter -> compact trace printer -> TODO/PLAN block with todo details
```

## Preconditions
- Native opencode `tool_use` events for `TodoWrite` carry todo items in
  `part.state.input.todos`.
- The trace printer should show the todo item details below the `TODO` header.

## Steps
1. Build one native opencode `tool_use` JSONL line for a completed
   `TodoWrite` call.
2. Include two todos with status and content in `part.state.input.todos`.
3. Format the line with `print.FormatTraceLine`.

```go
import (
	"encoding/json"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	line := map[string]any{
		"type":      "tool_use",
		"sessionID": "sess_todo",
		"part": map[string]any{
			"id":     "part_todo",
			"type":   "tool",
			"tool":   "TodoWrite",
			"callID": "call_todo",
			"state": map[string]any{
				"status": "completed",
				"input": map[string]any{
					"todos": []map[string]any{
						{"content": "Inspect pricing BFF routes", "status": "completed"},
						{"content": "Update credit.spl.bff glossary", "status": "pending"},
					},
				},
			},
		},
	}
	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("marshal opencode todo event: %v", err)
	}
	req.Line = string(data)
	return nil
}
```
