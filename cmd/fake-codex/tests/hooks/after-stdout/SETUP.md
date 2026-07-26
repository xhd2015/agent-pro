## Preconditions
- The mock config contains an `after_stdout` hook.

## Steps
1. Run fake Codex with the hook recorder.

```go
import (
    "fmt"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","hook_command":%q,"hooks":[{"at":"after_stdout","event":"AfterStdout","payload":{"ok":true}}],"llm_events":[{"type":"message","text":"before hook"}]}`, hook+" {{event}}"))
    return nil
}
```

