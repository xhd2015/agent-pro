## Preconditions
- The converter router receives raw runner-native JSONL and returns AgentEvent values.

## Steps
1. Set `req.Operation = "convert_raw"`.
2. Each leaf sets `req.AgentRunner` and `req.RawJSON`.
3. `Run` calls `convert.ConvertRawLine`, returns the AgentEvent JSON.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Operation = "convert_raw"
    return nil
}
```
