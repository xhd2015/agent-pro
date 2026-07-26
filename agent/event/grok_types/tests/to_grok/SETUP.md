## Preconditions
- Grouping node for ToGrok conversion tests (canonical AgentEvent → grok Event).

## Steps
1. Sets `Target` to `"to_grok"` as a default for all to_grok leaves.
2. Sets `SessionID` to a default test session ID.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "to_grok"
	req.SessionID = "sess_test_001"
	return nil
}
```
