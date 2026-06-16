## Preconditions
- Two JSON lines: a `PhaseUpdate` followed by a `PhaseEnd` for the same message ID.

## Steps
1. Feed `PhaseUpdate` JSON: `{"type":"message","phase":"update","id":"m1","text":"hello"}`
2. Feed `PhaseEnd` JSON: `{"type":"message","phase":"end","id":"m1","text":"hello world"}`
3. The `PhaseUpdate` should produce formatted output; the `PhaseEnd` should be suppressed.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Lines = []string{
		`{"type":"message","phase":"update","id":"m1","text":"hello"}`,
		`{"type":"message","phase":"end","id":"m1","text":"hello world"}`,
	}
	return nil
}
```
