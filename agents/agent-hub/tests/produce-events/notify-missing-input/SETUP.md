## Preconditions
- notify is called without --json or --file.

## Steps
1. Run `agent-hub notify` (no flags).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"notify"}
    return nil
}
```
