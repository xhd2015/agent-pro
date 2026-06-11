## Preconditions
- The mock config contains a simple completion event.
- `CODEX_THREAD_ID` is not set before the call.

## Steps
1. Write mock config with a single text event.
2. Run `doctest agent implement "implement the feature"` with `--agent-runner fake-codex`.
3. Verify exit code and session creation.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"implementation done","status":"completed"}}]}`)
    req.Args = []string{"--agent-runner", "fake-codex", "--mock-config", req.MockConfigPath, "implement the feature"}
    return nil
}
```
