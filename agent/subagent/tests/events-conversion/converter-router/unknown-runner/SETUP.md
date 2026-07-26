## Preconditions
- The converter router receives an unknown agent runner name.

## Steps
1. Set `req.AgentRunner = "unknown_runner_xyz"`.
2. Provide any raw JSON (e.g., an empty JSON object).
3. Call `convert.ConvertRawLine`.
4. Verify a non-nil error is returned.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.AgentRunner = "unknown_runner_xyz"
    req.RawJSON = `[{}]`
    return nil
}
```
