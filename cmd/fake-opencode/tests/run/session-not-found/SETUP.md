## Preconditions
- The command passes `--session no-such-session` but no session directory exists.

## Steps
1. Run fake opencode with `--session` pointing to a non-existent session.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","llm_events":[{"type":"message","text":"should not appear"}]}`)
    req.Args = []string{"run", "--format", "json", "--session", "no-such-session", "--mock-config", req.MockConfigPath, "hello"}
    return nil
}
```
