## Preconditions
- The mock config contains an `on_error` hook and a nonzero exit code.

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
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","exit_code":9,"stderr":"planned","hook_command":%q,"hooks":[{"at":"on_error","event":"OnError","payload":{"ok":true}}],"llm_events":[]}`, hook+" {{event}}"))
    return nil
}
```

