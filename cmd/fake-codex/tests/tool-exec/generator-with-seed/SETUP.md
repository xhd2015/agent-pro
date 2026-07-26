## Preconditions
- The random generator is used without mock config.
- A fixed seed ensures deterministic tool choices.

## Steps
1. Run fake-codex with `--seed 999888777` and a known prompt.
2. Verify the output has the expected structure: events contain type and item fields.
3. The tool choices are deterministic (same seed always produces same sequence).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"exec", "--json", "--seed", "999888777", "deterministic test prompt for seed check"}
    req.Env = append(req.Env, "CODEX_SANDBOX_SKIP=1")
    return nil
}
```
