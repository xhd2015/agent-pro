## Preconditions
- Grouping node for roundtrip conversion tests (ToGrok → FromGrok preserves key fields).

## Steps
1. Sets `Target` to `"roundtrip"` as a default for all roundtrip leaves.
2. Sets a default `SessionID`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.SessionID = "sess_roundtrip_001"
	return nil
}
```
