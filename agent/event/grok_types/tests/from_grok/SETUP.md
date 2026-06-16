## Preconditions
- Grouping node for FromGrok conversion tests (grok Event → canonical AgentEvent).

## Steps
1. Sets `Target` to `"from_grok"` as a default for all from_grok leaves.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_grok"
	return nil
}
```
