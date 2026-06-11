## Preconditions
- The mock config does NOT have `session_id` set.
- fake-codex auto-generates a `CODEX_THREAD_ID`.

## Steps
1. Write mock config without `session_id`, with a hook that captures the env var.
2. Run fake-codex.
3. Read the marker and verify a non-empty CODEX_THREAD_ID was set.

```go
import (
    "fmt"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    marker := filepath.Join(req.TempDir, "thread-id.txt")
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","hook_command":"echo $CODEX_THREAD_ID > %s","hooks":[{"at":"before_start","event":"capture","payload":{}}],"stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"done","status":"completed"}}]}`, marker))
    return nil
}
```
