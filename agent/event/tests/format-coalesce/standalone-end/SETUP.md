# Scenario

**Feature**: A single `PhaseEnd` JSON line with no prior delta events for the same ID

## Preconditions
- A single `PhaseEnd` JSON line with no prior delta events for the same ID.

## Steps
1. Feed one raw JSON line: `{"type":"message","phase":"end","id":"m2","text":"done"}`
2. Since there is no prior event for ID "m2", it must produce formatted output.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Lines = []string{
		`{"type":"message","phase":"end","id":"m2","text":"done"}`,
	}
	return nil
}
```
