## Preconditions
- `CODEX_THREAD_ID` is set to simulate a prior session.

## Steps
1. Set `CODEX_THREAD_ID`.
2. Write mock config with completion event.
3. Run `doctest agent implement` with answers.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "CODEX_THREAD_ID=impl_test_session_1")
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","session_id":"impl_test_session_1","stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"resumed and done","status":"completed"}}]}`)
    req.Args = []string{"agent", "implement", "--agent-runner", "fake-codex", "the port should be 8080"}
    return nil
}
```
