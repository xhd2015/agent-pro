## Preconditions
- agent_start has no payload fields beyond `type`.

## Steps
1. Parse the agent_start event JSON.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"agent_start"}`
	return nil
}
```
