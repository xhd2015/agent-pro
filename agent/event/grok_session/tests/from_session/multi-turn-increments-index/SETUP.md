# Scenario

**Feature**: turn_index increments across turns

```
turn0(user+done) -> turn_index=0; turn1(user+done) -> turn_index=1
```

## Preconditions
- Two turns separated by turn_completed boundaries.

## Steps
1. Provide two user+turn_completed sequences.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("first turn"),
		acpTurnCompleted(),
		acpUserChunk("second turn"),
		acpTurnCompleted(),
	}
	return nil
}
```
