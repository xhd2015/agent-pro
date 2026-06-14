## Preconditions
- The mock config contains a `command_execution` event with no `"mock"` object.

## Steps
1. Write a mock config with a command_execution event that echoes a known string.
2. Run fake-codex with `--json`.
3. Verify the event contains the real command output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","session_id":"codex_cmd_real","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo hello real codex"}}]}`)
    return nil
}
```
