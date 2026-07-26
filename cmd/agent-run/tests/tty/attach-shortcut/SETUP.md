# Scenario

**Bug**: `agent-run attach <session-id>` is an alias for `agent-run tty attach <session-id>`

```
agent-run attach <id> -> delegates to same logic as tty attach -> same error output
agent-run tty attach <id> -> delegates to same logic -> same error output
```

## Preconditions

- Both `attach` and `tty attach` delegate to the same implementation.
- An error produced by one must match the error from the other when given the same input.

## Steps

1. `Setup` configures a bad session id (no registry entry).
2. Paired leaves run `attach <bad-id>` and `tty attach <bad-id>` and compare errors.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RegistryDir = "grok-tty-registry"
	req.RegistrySessionID = "session-nonexistent"
	return nil
}
```
