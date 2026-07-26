# Scenario

**Feature**: two-turn session roundtrips

```
two turns with turn_index 0/1 and two ActionDone
```

## Preconditions
- Two complete turns in wire input.

## Steps
1. Seed two user+assistant+turn_completed sequences.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("first turn"),
		acpAssistantChunk("answer one"),
		acpTurnCompleted(),
		acpUserChunk("second turn"),
		acpAssistantChunk("answer two"),
		acpTurnCompleted(),
	}
	return nil
}
```
