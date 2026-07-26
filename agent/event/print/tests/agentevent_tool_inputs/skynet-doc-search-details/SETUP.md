# Scenario

**Bug**: MCP `skynet-base_get_doc_content` tool calls print only a bare header

```
# maintain-topic uses skynet MCP to search Confluence for pricing-center docs
AgentEvent{tool=skynet-base_get_doc_content, tool_input.doc=search URL} -> compact trace printer

# compact output should include the search/doc URL below the tool header
compact trace printer -> tool block with doc/search target
```

## Preconditions
- The agent invoked `skynet-base_get_doc_content` with a Confluence search URL
  while looking for `credit.pricing.center` documentation.
- This is the "web search" step the user sees as a bare tool header in `--print`.

## Steps
1. Build one canonical AgentEvent JSONL line for the skynet doc lookup.
2. Include the Confluence search URL in `tool_input.doc`.
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
		Tool: "skynet-base_get_doc_content",
		ToolInput: map[string]any{
			"doc": "https://fake.xhd2015.xyz/search?text=credit+pricing+center",
		},
	})
	if err != nil {
		t.Fatalf("marshal skynet doc search AgentEvent: %v", err)
	}
	req.Line = string(data)
	return nil
}
```