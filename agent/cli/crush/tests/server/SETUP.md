## Preconditions
- All tests under this directory use server-mode (`Mode="convert"`, `Mode="server-client"`, or `Mode="server-ask"`).
- `convert` sub-tests are pure unit tests — no server needed.
- `server-client` and `session-persist` sub-tests require `crush` binary in PATH.

## Steps
1. Set default mode to `"server-ask"` — child Setups may override.
2. Leaves under `convert/` override to `"convert"` and set `SSEInput`.
3. Leaves under `server-client/` override to `"server-client"` and set `ServerOperation`.
4. `session-persist` leaf inherits the default `"server-ask"` mode.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "server-ask"
	return nil
}
```
