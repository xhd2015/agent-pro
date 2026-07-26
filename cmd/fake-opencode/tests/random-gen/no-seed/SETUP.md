## Preconditions
- No seed is provided; a random seed is generated internally.

## Steps
1. Run fake opencode without a mock config or seed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"run", "--format", "json", "hello world"}
    return nil
}
```
