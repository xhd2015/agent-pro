## Preconditions
- The mock config contains a `before_stdout` hook and stdout marker.

## Steps
1. Run fake Codex with the hook recorder.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","hook_command":%q,"hooks":[{"at":"before_stdout","event":"BeforeStdout","payload":{"ok":true}}],"llm_events":[{"type":"message","text":"after hook"}]}`, hook+" {{event}}"))
    return nil
}
```

