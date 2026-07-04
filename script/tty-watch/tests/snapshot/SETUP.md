# Scenario

**Feature**: `tty-watch snapshot` fetches scrollback with sanitized print output

```
# snapshot mode via WS
tty-watch snapshot <id> -> attach_mode=snapshot -> sanitize ANSI/C0 -> stdout
```

## Preconditions

- `prints-sanitized` starts a detached session that emits ANSI-colored text.
- `unknown-session` uses a nonexistent session id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("snapshot setup: tty-watch binary not built")
	}
	return nil
}
```