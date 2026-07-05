# Scenario

**Feature**: turn_index equal in events₁ vs events₂

```
multi-event turn -> roundtrip preserves per-event turn_index
```

## Preconditions
- Multi-event turn with explicit turn_index stamping.

## Steps
1. Seed full turn; compare turn_index field-wise across passes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("persist events"),
		acpThoughtChunk("planning"),
		acpAssistantChunk("done"),
		acpTurnCompleted(),
	}
	return nil
}
```
