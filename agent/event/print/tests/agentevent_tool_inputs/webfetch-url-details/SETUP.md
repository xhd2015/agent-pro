# Scenario

**Bug**: canonical `webfetch` AgentEvent tool calls print only the `WEBFETCH` header

```
# maintain-topic records Confluence page fetches as canonical AgentEvent tool_call lines
AgentEvent{tool=webfetch, tool_input.url=...} -> compact trace printer

# compact output should include the fetched URL below the WEBFETCH header
compact trace printer -> WEBFETCH block with target URL
```

## Preconditions
- A `webfetch` tool call carries the target URL in `tool_input.url`.
- Session `20260625-114542-...-credit.pricing.center` fetched two Confluence pages.

## Steps
1. Build one canonical AgentEvent JSONL line for a completed `webfetch` tool call.
2. Include the Confluence page URL in `tool_input.url`.
3. Format the line with `print.FormatTraceLine`.

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	data, err := json.Marshal(types.AgentEvent{
		Type: types.ActionToolCall,
		Tool: "webfetch",
		ToolInput: map[string]any{
			"url": "https://fake.xhd2015.xyz/pages/viewpage.action?pageId=830343951",
		},
	})
	if err != nil {
		t.Fatalf("marshal webfetch AgentEvent: %v", err)
	}
	req.Line = string(data)
	return nil
}
```