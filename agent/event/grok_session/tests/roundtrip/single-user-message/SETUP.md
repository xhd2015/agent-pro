# Scenario

**Feature**: single user message roundtrips with turn_index

```
user_message_chunk + turn_completed -> events -> wire -> events (semantic equal)
```

## Preconditions
- One user turn with turn_completed boundary.

## Steps
1. Seed user chunk and turn_completed; run roundtrip harness.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("run ls"),
		acpTurnCompleted(),
	}
	return nil
}
```
