# Scenario

**Feature**: A3 — busy `--no-wait` returns fast; later idle deliver via TTY drainer (no blocking send)

```
busy-then-idle stub (~12s busy frames) ServeSession (drainer waiting for writable)
  -> agent-run send <id> "tty-drainer-busy-probe" --no-wait
  -> <1s, stdout msg_N, exit 0 (CLI gone)
  -> screen becomes idle → session drainer injects probe
  -> msg status .../msg_N -> delivered
```

## Steps

1. Set `req.Action = "busy-no-wait-delivers-when-idle"`.
2. Set `req.SendMessage = "tty-drainer-busy-probe"`.
3. Set `req.ExecTimeout = 35s` to cover busy→idle transition + inject.

```go
import "time"

func Setup(t *testing.T, req *Request) error {
	req.Action = "busy-no-wait-delivers-when-idle"
	req.SendMessage = "tty-drainer-busy-probe"
	req.ExecTimeout = 35 * time.Second
	return nil
}
```
