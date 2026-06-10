## Preconditions
- The envelope has a negative offset.

## Steps
1. Select negative offset.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "envelope-negative-offset"; return nil }
```

