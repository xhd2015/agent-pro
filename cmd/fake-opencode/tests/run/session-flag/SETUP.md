## Preconditions
- The command passes `--session sess_arg`.

## Steps
1. Run fake opencode with a mock config that does not define a session.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","stdout_events":[{"type":"text","part":{"id":"p1","type":"text","text":"session flag"}}]}`)
    req.Args = []string{"run", "--format", "json", "--session", "sess_arg", "--mock-config", req.MockConfigPath, "hello"}
    return nil
}
```

