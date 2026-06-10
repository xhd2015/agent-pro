## Preconditions
- Valid event JSON is read from a file.
## Steps
1. Notify from file.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "notify-file-valid"; return nil }
```

