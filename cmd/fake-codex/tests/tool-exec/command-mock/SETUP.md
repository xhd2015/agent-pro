## Preconditions
- The mock config contains a `command_execution` event **with** a `"mock"` object.

## Steps
1. Write a mock config with `"mock":{"output":"fake codex output","exit_code":0}`.
2. The command would produce different output if actually executed.
3. Run fake-codex and verify mock output is used.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","session_id":"codex_cmd_mock","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo REAL CODEX OUTPUT"},"mock":{"output":"fake codex output","exit_code":0}}]}`)
    return nil
}
```
