## Preconditions
- `showStatus` reads events.jsonl and displays status summary.

## Steps
1. Set `req.Operation = "status"`.
2. Pre-create session directories with events.jsonl.
3. `Run` calls `subagent.TestExported_showStatus` with `Status=true`.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Operation = "status"
    return nil
}
```
