## Preconditions
- The mock config contains session and model metadata.
- A hook recorder captures the hook payload.

## Steps
1. Run fake Codex and fire a `SessionStart` hook.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","session_id":"sess_meta","model":"gpt-test","hook_command":%q,"hooks":[{"at":"before_stdout","event":"SessionStart","payload":{"source":"startup"}}],"llm_events":[]}`, hook+" {{event}}"))
    return nil
}
```

