## Preconditions
- A consumer fetches a two-event batch.

## Steps
1. Load the cursor after fetch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "cursor-advance-after-batch"; return nil }
```

