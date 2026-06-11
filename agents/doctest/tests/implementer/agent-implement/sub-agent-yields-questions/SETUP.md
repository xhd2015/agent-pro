## Preconditions
- The mock config has a `before_exit` hook that calls `yield-pending-questions`.

## Steps
1. Write mock config with a hook.
2. Run `doctest agent implement "implement feature"`.
3. Questions are printed to stdout.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    yieldPQ := ""
    for _, env := range req.Env {
        if len(env) > 13 && env[:13] == "YIELD_PQ_BIN=" {
            yieldPQ = env[13:]
            break
        }
    }
    if yieldPQ == "" {
        t.Fatal("YIELD_PQ_BIN not set by parent")
    }

    hookCmd := yieldPQ + " '{\"id\":\"1\",\"question\":\"What is the target port?\"}'"
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","hook_command":%q,"hooks":[{"at":"before_exit","event":"yield","payload":{"ok":true}}],"stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"working on it","status":"completed"}}]}`, hookCmd+" {{event}}"))
    req.Args = []string{"agent", "implement", "--agent-runner", "fake-codex", "implement feature"}
    return nil
}
```
