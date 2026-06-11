## Preconditions
- Neither `DOCTEST_AGENT_IMPLEMENTER_SESSION_ID` nor `CODEX_THREAD_ID` is set.
- Auto-discovery from parent opencode process may or may not succeed.

## Steps
1. Set no session ID env vars.
2. Write a mock config for fake-codex.
3. Run `doctest agent implement "test" --agent-runner fake-codex`.
4. ASSERT handles the conditional: skip if no opencode ancestor, otherwise verify.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{
        "version":"agent-pro.fake-runner.v1",
        "runner":"fake-codex",
        "session_id":"inner-from-discovery",
        "stdout_events":[
            {"type":"item.completed","item":{"id":"m1","type":"message","text":"done","status":"completed"}}
        ]
    }`)

    req.Args = []string{"agent", "implement", "--agent-runner", "fake-codex", "test"}
    return nil
}
```
