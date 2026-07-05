# Scenario

**Feature**: assistant message roundtrips

```
agent_message_chunk + turn_completed -> semantic equal events
```

## Preconditions
- Assistant chunk within a turn.

## Steps
1. Seed user + assistant + turn_completed for turn context.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("prompt"),
		acpAssistantChunk("Here is the answer"),
		acpTurnCompleted(),
	}
	return nil
}
```
