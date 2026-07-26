## Preconditions
- `traceSession` reads events.jsonl and displays formatted output.

## Steps
1. Set `req.Operation = "trace"`.
2. Pre-create session directories with events.jsonl.
3. `Run` calls `subagent.TestExported_traceSession` with `CatchUp=true`.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Operation = "trace"
    return nil
}
```
