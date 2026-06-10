## Preconditions
- Consumer cursor exists.
## Steps
1. Replay from explicit cursor.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "replay-from-cursor"; return nil }
```

