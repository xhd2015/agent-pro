## Preconditions
- The `agent-hub` binary is built (inherited from root SETUP.md).
- `--help` and `-h` flags trigger help output with exit code 0.

## Steps
1. Run `agent-hub <command> --help` (or `-h`) for each command level.
2. Verify stdout contains the expected help content: usage line, command list with descriptions, and flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    _ = t
    return nil
}
```
