# Scenario

**Feature**: flat user_message_chunk converts to user ActionMessage with turn_index=0

```
# single user chunk coalesces and flushes at end of input
user_message_chunk{"text":"run ls"} -> ActionMessage role=user turn_index=0
```

## Preconditions
- One flat `user_message_chunk` line with text `run ls`.

## Steps
1. Provide one `acpUserChunk("run ls")` wire line.
2. Call `FromUpdatesJSONL`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{acpUserChunk("run ls")}
	return nil
}
```
