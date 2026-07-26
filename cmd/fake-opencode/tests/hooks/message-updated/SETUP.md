## Preconditions
- The mock config fires `message.updated`.

## Steps
1. Run fake opencode with a hook recorder.

```go
import (
    "fmt"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    hook := writeHookRecorder(t, req, 0)
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","hook_command":%q,"hooks":[{"at":"before_stdout","event":"message.updated","payload":{"message":{"role":"user","text":"prompt text"}}}],"llm_events":[]}`, hook+" {{event}}"))
    return nil
}
```

