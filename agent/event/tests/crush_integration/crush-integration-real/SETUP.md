## Preconditions
- `crush` is available on `PATH` (or skip via `CRUSH_SKIP_INTEGRATION=1`).
- A free TCP port is available for the server.

## Steps
1. Use auto-assigned port (`HostPort = 0`).
2. Use default `CrushPath` (empty → `LookPath`).
3. Send the prompt `"one word of French capital"`.

## Context
- The root `Run` function auto-locates the crush binary and picks a free port.
- If the binary is not found the test is skipped.
- The prompt expects a single-word answer containing "paris".

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "one word of French capital"
	req.HostPort = 0
	return nil
}
```
