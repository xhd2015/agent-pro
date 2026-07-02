# Scenario

**Bug**: follow-up message in a web tty chat must reuse the mapped PTY

```
finished chat web_* + terminal_session_id session-1 + live registry
POST /messages second prompt -> same terminal_session_id session-1
```

## Preconditions

- Parent setup created one live mapped PTY.
- The PTY is reachable, so replacement is not allowed.

## Steps

1. Probe terminal status before follow-up.
2. POST a follow-up message to the same web chat route.
3. Probe terminal status after follow-up.
4. List registry ids under `codex-tty-registry`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "followup"
	return nil
}
```
