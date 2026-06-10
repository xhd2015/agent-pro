## Preconditions
- The test runs `fake-opencode models`.

## Steps
1. Run the models command.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"models"}
    return nil
}
```

