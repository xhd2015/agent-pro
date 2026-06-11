## Preconditions
- The mock config fires `session.created`.

## Steps
1. Run fake opencode with a hook recorder.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_hook","hook_command":%q,"hooks":[{"at":"before_stdout","event":"session.created","payload":{"ok":true}}],"stdout_events":[]}`, hook+" {{event}}"))
    return nil
}
```

