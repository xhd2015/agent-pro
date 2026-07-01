# Scenario

**Feature**: ActionMessage maps to an assistant text block

```
# canonical message event -> assistant text content block
ActionMessage(text="Hello") -> {"role":"assistant","content":[{"type":"text","text":"Hello"}]}
```

## Preconditions
- `ToCmd` converts each `ActionMessage` into an assistant event with a `text` content block.

## Steps
1. Provide one canonical `ActionMessage` event.
2. Verify the output contains an assistant event with a `text` block.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, Text: "Hello"},
	}
	return nil
}
```
