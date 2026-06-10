## Preconditions
- This branch tests Codex hook events.
## Steps
1. Mark branch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "codex" }; return nil }
```

