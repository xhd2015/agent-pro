## Preconditions
- `bun` is installed and available in PATH (test is skipped with a warning otherwise).
- The plugin file is a TypeScript module exporting event handlers.
- fake-opencode supports `--plugin <path>` and `mockConfig.plugins`.

## Steps
1. Build `cmd/fake-opencode`.
2. Write a plugin `.ts` file that logs handler invocations.
3. Run fake-opencode with `--plugin` or `mockConfig.plugins`.
4. Verify handler was invoked for matching events.

```go
import (
    "os/exec"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    if _, err := exec.LookPath("bun"); err != nil {
        t.Skipf("skipping plugin test: bun not installed")
    }
    return nil
}
```
