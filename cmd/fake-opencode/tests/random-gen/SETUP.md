## Preconditions
- No mock config is provided.
- fake-opencode falls back to random event generation using the neutral AgentEvent type.
- Random generation is seeded for determinism.

## Steps
1. Mark the test mode as random-gen.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=random-gen")
    return nil
}
```
