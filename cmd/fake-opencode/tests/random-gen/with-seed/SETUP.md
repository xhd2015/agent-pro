## Preconditions
- A fixed seed is provided via `--seed` flag.
- Random generation produces the same output on every run.

## Steps
1. Run fake opencode without a mock config, using a fixed seed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"run", "--format", "json", "--seed", "42", "hello world"}
    return nil
}
```
