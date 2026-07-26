# Scenario

**Feature**: envelope wire roundtrips semantically

```
envelope user chunk -> events -> wire -> events
```

## Preconditions
- Input uses `_x.ai/session/update` envelope lines.

## Steps
1. Seed envelope user chunk + turn_completed.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpEnvelopeUserChunk(req.SessionID, "run ls"),
		acpTurnCompleted(),
	}
	return nil
}
```
