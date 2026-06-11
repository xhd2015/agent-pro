## Preconditions
- The mock config has a `before_exit` hook that calls `yield-pending-questions`.
- The `yield-pending-questions` binary is in the PATH (copied from doctest).

## Steps
1. Write mock config with a hook that calls yield-pending-questions.
2. Run `doctest agent implement "implement feature"`.
3. The hook fires yield-pending-questions, which writes to a file.
4. `doctest agent implement` reads the file and prints the questions to stdout.

```go
import (
    "fmt"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    hookCmd := req.YieldPQBin + " '{\"id\":\"1\",\"question\":\"What is the target port?\"}'"
    writeMockConfig(t, req, fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","hook_command":%q,"hooks":[{"at":"before_exit","event":"yield","payload":{"ok":true}}],"stdout_events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"working on it","status":"completed"}}]}`, hookCmd+" {{event}}"))

    req.Args = []string{"--agent-runner", "fake-codex", "--mock-config", req.MockConfigPath, "implement feature"}
    return nil
}
```
