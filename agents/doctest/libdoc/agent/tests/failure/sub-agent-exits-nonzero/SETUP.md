## Preconditions
- The mock config has `exit_code: 3` and stderr text.

## Steps
1. Write mock config with non-zero exit.
2. Run `doctest agent implement "implement feature"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","exit_code":3,"stderr":"sub-agent implementation failed","stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"partial work","status":"completed"}}]}`)
    req.Args = []string{"--agent-runner", "fake-codex", "--mock-config", req.MockConfigPath, "implement feature"}
    return nil
}
```
