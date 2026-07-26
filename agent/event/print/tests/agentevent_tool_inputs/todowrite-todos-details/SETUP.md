# Scenario

**Bug**: canonical `todowrite` AgentEvent tool calls print only the `TODO` header

```
# maintain-topic records plan updates as canonical AgentEvent tool_call lines
AgentEvent{tool=todowrite, tool_input.todos[]} -> compact trace printer

# compact output should include todo content and status below the TODO header
compact trace printer -> TODO block with plan item details
```

## Preconditions
- A `todowrite` tool call carries todo items in `tool_input.todos`.
- Each todo has `content`, `status`, and optionally `priority`.

## Steps
1. Build one canonical AgentEvent JSONL line for a completed `todowrite` tool call.
2. Include todos from the maintain-topic session in `tool_input.todos`.
3. Format the line with `print.FormatTraceLine`.

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	data, err := json.Marshal(types.AgentEvent{
		Type: types.ActionToolCall,
		Tool: "todowrite",
		ToolInput: map[string]any{
			"todos": []map[string]any{
				{
					"content":  "搜索 Confluence 上 credit.pricing.center 相关文档",
					"priority": "high",
					"status":   "in_progress",
				},
				{
					"content":  "搜索 git 仓库了解项目信息",
					"priority": "high",
					"status":   "pending",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal todowrite AgentEvent: %v", err)
	}
	req.Line = string(data)
	return nil
}
```