## Preconditions
- The mock config exits nonzero and fires `session.error`.

## Steps
1. Run fake opencode with a hook recorder.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","exit_code":6,"stderr":"bad session","hook_command":%q,"hooks":[{"at":"on_error","event":"session.error","payload":{"error":"bad session"}}],"stdout_events":[]}`, hook+" {{event}}"))
    return nil
}
```

