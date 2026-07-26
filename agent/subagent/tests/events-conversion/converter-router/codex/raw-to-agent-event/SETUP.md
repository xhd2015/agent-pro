## Preconditions
- The converter router receives codex-native JSON events (item.started + item.completed).

## Steps
1. Set `req.AgentRunner = "codex"`.
2. Feed codex events: command_execution started+completed, message completed.
3. Call `convert.ConvertRawLine`.
4. Verify AgentEvent output.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.AgentRunner = "codex"
    req.RawJSON = `[{"type":"item.started","item":{"id":"c1","type":"command_execution"}},{"type":"item.completed","item":{"id":"c1","type":"command_execution","command":"ls","aggregated_output":"file1.txt\nfile2.txt","exit_code":0,"status":"completed"}},{"type":"item.completed","item":{"id":"m1","type":"message","text":"Codex says hello","status":"completed"}}]`
    return nil
}
```
