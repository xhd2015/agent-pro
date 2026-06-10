## Preconditions
- Invalid JSON is passed inline.
## Steps
1. Notify invalid JSON.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "notify-json-invalid"; return nil }
```

