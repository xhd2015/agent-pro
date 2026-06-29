## Preconditions
- `crush` is available on `PATH` (or skip via `CRUSH_SKIP_INTEGRATION=1`).
- A free TCP port is available for the server.

## Steps
1. Use auto-assigned port (`HostPort = 0`).
2. Use default `CrushPath` (empty → `LookPath`).
3. Send the prompt `"What is the capital city of France? Answer with exactly one word."`.

## Context
- The root `Run` function auto-locates the crush binary and picks a free port.
- If the binary is not found the test is skipped.
- The prompt expects a single-word answer containing "paris".

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "What is the capital city of France? Answer with exactly one word."
	req.HostPort = 0
	return nil
}
```
