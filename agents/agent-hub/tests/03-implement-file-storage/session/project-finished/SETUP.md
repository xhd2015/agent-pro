## Preconditions
- Started and finished events are appended.

## Steps
1. Read the completed session projection.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "session-project-finished"; return nil }
```

