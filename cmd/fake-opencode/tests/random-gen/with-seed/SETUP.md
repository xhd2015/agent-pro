## Preconditions
- A fixed seed is provided via `--seed` flag.
- Random generation produces the same output on every run.

## Steps
1. Run fake opencode without a mock config, using a fixed seed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"run", "--format", "json", "--seed", "42", "hello world"}
    return nil
}
```
