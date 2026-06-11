## Preconditions
- `CODEX_THREAD_ID` is set from a previous call.
- A followup message is provided with answers.

## Steps
1. Set `CODEX_THREAD_ID` to simulate a prior session.
2. Write mock config with a completion event.
3. Run `doctest agent implement` with answers as the prompt.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "CODEX_THREAD_ID=impl_test_session_1")

    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","session_id":"impl_test_session_1","stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"resumed and done","status":"completed"}}]}`)
    req.Args = []string{"--agent-runner", "fake-codex", "--mock-config", req.MockConfigPath, "the port should be 8080"}
    return nil
}
```
