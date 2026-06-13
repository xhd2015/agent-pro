## Steps
1. Run `agent-hub integration install opencode` (local, no flags).

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "install", "opencode"}
    return nil
}
```
