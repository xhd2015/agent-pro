# Scenario

**Feature**: ActionThink maps to an assistant reasoning block

```
# canonical think event -> assistant reasoning content block
ActionThink(text="thinking...") -> {"role":"assistant","content":[{"type":"reasoning","text":"thinking..."}]}
```

## Preconditions
- `ToCmd` converts each `ActionThink` into an assistant event with a `reasoning` content block.

## Steps
1. Provide one canonical `ActionThink` event.
2. Verify the output contains an assistant event with a `reasoning` block.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionThink, Text: "thinking..."},
	}
	return nil
}
```
