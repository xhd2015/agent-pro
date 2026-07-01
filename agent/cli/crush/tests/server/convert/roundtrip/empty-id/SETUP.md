# Scenario

**Feature**: ToCrush/FromCrush assigns a synthetic ID for a message event with an empty ID

## Preconditions
- Input is a message `AgentEvent` with an empty `ID` field.
- `ToCrush` should assign a synthetic ID (`"evt_1"`).
- `FromCrush` prepends `"crush:"`, resulting in `"crush:evt_1"`.

## Steps
1. Set `AgentEventsJSON` to a message event without an ID.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AgentEventsJSON = `[{"type":"message","text":"test","id":""}]`
	return nil
}
```
