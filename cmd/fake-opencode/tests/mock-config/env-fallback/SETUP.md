## Preconditions
- The mock config path is provided through `FAKE_OPENCODE_MOCK_CONFIG`.

## Steps
1. Run fake opencode without `--mock-config`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_env","stdout_events":[{"type":"message","text":"from env"}]}`)
    req.Env = append(req.Env, "FAKE_OPENCODE_MOCK_CONFIG="+req.MockConfigPath)
    req.Args = []string{"run", "--format", "json", "hello"}
    return nil
}
```
