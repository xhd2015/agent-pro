## Preconditions
- No existing plugin file.
- Local project directory is the current working directory.

## Steps
1. Run `agent-hub integration install opencode`.
2. Verify plugin file was created under `.opencode/plugins/agent-hub.ts`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    _ = t
    return nil
}
```
