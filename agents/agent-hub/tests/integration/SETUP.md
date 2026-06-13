## Preconditions
- The `agent-hub` binary is built (inherited from root SETUP.md).
- Integration commands manage opencode plugins via file operations in `.opencode/plugins/` (local) or `~/.config/opencode/plugins/` (global).

## Steps
1. Run `agent-hub integration <subcommand>` with appropriate arguments.
2. Verify exit code, stdout message, and file-system side effects.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    _ = t
    return nil
}
```
