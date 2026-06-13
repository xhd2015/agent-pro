## Preconditions
- --from argument has bad format.

## Steps
1. Run `agent-hub replay --consumer-id c1 --from bogus`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"replay", "--consumer-id", "c1", "--from", "bogus"}
    return nil
}
```
