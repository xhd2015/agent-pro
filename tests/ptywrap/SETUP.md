# Scenario

**Feature**: ptywrap server exposes PTY sessions over HTTP and WebSocket

```
# REST create / list / rename
test client -> POST/PATCH/GET /api/terminal/sessions -> session manager -> PTY child

# WS attach
test client -> WS /api/terminal -> session attach -> PTY I/O + scrollback
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` is importable.
- `ptywrap.RegisterAPI` mounts handlers on a test `http.ServeMux`.
- `python3` available on PATH for TTY probe leaf (or implementer substitutes equivalent).

## Steps

1. Start ephemeral HTTP test server with `ptywrap.RegisterAPI`.
2. Set `req.ServerBase` to the listener URL shared by all leaves.

## Context

- WS clients use `gorilla/websocket` DefaultDialer.
- Scrollback replay timing may need tuning in implementer pass.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	base, cleanup := startTestServer(t)
	t.Cleanup(cleanup)
	req.ServerBase = base
	return nil
}
```