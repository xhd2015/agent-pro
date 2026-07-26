## Preconditions
- The mock config has `session_id` set.
- A hook captures `CODEX_THREAD_ID` from the environment.

## Steps
1. Write mock config with `session_id` and a `before_start` hook.
2. The hook writes `$CODEX_THREAD_ID` to a marker file.
3. Run fake-codex.
4. Read the marker file and verify it matches the session_id.

```go
import (
    "fmt"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    marker := filepath.Join(req.TempDir, "thread-id.txt")
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","session_id":"sess_thread_test","hook_command":"echo $CODEX_THREAD_ID > %s","hooks":[{"at":"before_start","event":"capture","payload":{}}],"llm_events":[{"type":"message","text":"done"}]}`, marker))
    return nil
}
```
