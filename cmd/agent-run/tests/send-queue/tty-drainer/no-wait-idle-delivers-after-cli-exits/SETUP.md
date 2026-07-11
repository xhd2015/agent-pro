# Scenario

**Feature**: A1 — idle stub-tty; `--no-wait` only; CLI fully exits; message still delivers

```
idle stub-tty ServeSession (drainer running)
  -> agent-run send <id> "tty-drainer-idle-probe" --no-wait
  -> stdout msg_N\n, exit 0, process exits
  -> (no further CLI send) session drainer injects probe into scrollback
  -> msg status .../msg_N -> delivered; queue lacks id
```

## Steps

1. Set `req.Action = "no-wait-idle-delivers-after-cli-exits"`.
2. Set `req.SendMessage = "tty-drainer-idle-probe"`.
3. Set `req.ExecTimeout = 30s` for inject poll after CLI exit.

```go
import "time"

func Setup(t *testing.T, req *Request) error {
	req.Action = "no-wait-idle-delivers-after-cli-exits"
	req.SendMessage = "tty-drainer-idle-probe"
	req.ExecTimeout = 30 * time.Second
	return nil
}
```
