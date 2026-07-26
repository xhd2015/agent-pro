## Preconditions
- No mock config is provided; the random generator is used.
- A prompt is provided that the generator will scan for tool suggestions.

## Steps
1. Create a temp directory with known files so real tool execution produces meaningful output.
2. Run fake-codex with a prompt containing `echo` and `ls` to trigger tool probes.
3. Verify the output does NOT contain old hardcoded fake data strings.

```go
import (
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Override default args: no mock config, use --seed for determinism
    req.Args = []string{"exec", "--json", "--seed", "12345", "echo GEN_TEST_MARKER && ls -la"}
    req.Env = append(req.Env, "CODEX_SANDBOX_SKIP=1")
    return nil
}
```
