# Scenario

**Feature**: Three JSON lines: `PhaseEnd`(m1), `ActionThink`, `PhaseEnd`(m2)

## Preconditions
- Three JSON lines: `PhaseEnd`(m1), `ActionThink`, `PhaseEnd`(m2).
- The think event (non-ActionMessage) should reset the coalescer, so both PhaseEnd events produce output.

## Steps
1. Feed `PhaseEnd` JSON for message m1: `{"type":"message","phase":"end","id":"m1","text":"first"}`
2. Feed think JSON: `{"type":"think","text":"hmm"}`
3. Feed `PhaseEnd` JSON for message m2: `{"type":"message","phase":"end","id":"m2","text":"second"}`
4. All three should produce formatted output (think resets message state).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Lines = []string{
		`{"type":"message","phase":"end","id":"m1","text":"first"}`,
		`{"type":"think","text":"hmm"}`,
		`{"type":"message","phase":"end","id":"m2","text":"second"}`,
	}
	return nil
}
```
