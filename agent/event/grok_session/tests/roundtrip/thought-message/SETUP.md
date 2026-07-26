# Scenario

**Feature**: thought chunk roundtrips

```
agent_thought_chunk -> ActionThink preserved semantically
```

## Preconditions
- Thought chunk in a turn.

## Steps
1. Seed user + thought + turn_completed.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("prompt"),
		acpThoughtChunk("planning ls output"),
		acpTurnCompleted(),
	}
	return nil
}
```
