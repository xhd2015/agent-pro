## Preconditions
- The command passes `--session no-such-session` but no session directory exists.

## Steps
1. Run fake opencode with `--session` pointing to a non-existent session.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","stdout_events":[{"type":"text","part":{"id":"p1","type":"text","text":"should not appear"}}]}`)
    req.Args = []string{"run", "--format", "json", "--session", "no-such-session", "--mock-config", req.MockConfigPath, "hello"}
    return nil
}
```
