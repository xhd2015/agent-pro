# Scenario

**Feature**: complete single turn roundtrips

```
user + think + tool + assistant + turn_completed -> semantic equal
```

## Preconditions
- Full-turn wire sequence.

## Steps
1. Seed full-turn lines and run roundtrip.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("persist events"),
		acpThoughtChunk("planning ls output"),
		acpToolCall("call_persist", "execute", "ls"),
		acpToolCallUpdate("call_persist", "completed", "agent\nagents"),
		acpAssistantChunk("PERSIST_ASSISTANT_MARKER"),
		acpTurnCompleted(),
	}
	return nil
}
```
