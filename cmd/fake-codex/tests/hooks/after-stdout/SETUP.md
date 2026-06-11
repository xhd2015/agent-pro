## Preconditions
- The mock config contains an `after_stdout` hook.

## Steps
1. Run fake Codex with the hook recorder.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","hook_command":%q,"hooks":[{"at":"after_stdout","event":"AfterStdout","payload":{"ok":true}}],"stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"before hook","status":"completed"}}]}`, hook+" {{event}}"))
    return nil
}
```

