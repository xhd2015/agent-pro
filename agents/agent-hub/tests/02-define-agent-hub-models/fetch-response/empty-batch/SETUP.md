## Preconditions
- The response contains an empty event batch.

## Steps
1. Select empty batch response.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "fetch-response-empty-batch"; return nil }
```

