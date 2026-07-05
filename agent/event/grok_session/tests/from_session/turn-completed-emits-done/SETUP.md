# Scenario

**Feature**: turn_completed emits ActionDone with turn_index

```
user_message_chunk -> turn_completed -> ActionDone turn_index=0
```

## Preconditions
- User chunk followed by `turn_completed`.

## Steps
1. Provide user chunk and turn_completed lines.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("hello"),
		acpTurnCompleted(),
	}
	return nil
}
```
