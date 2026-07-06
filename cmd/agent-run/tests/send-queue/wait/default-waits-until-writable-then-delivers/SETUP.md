# Scenario

**Feature**: default send waits past old 10s cap until terminal writable

```
busy screen 12s -> idle prompt -> default send blocks >10s -> exit 0, delivered
```

## Steps

1. Set `req.Action = "default-waits-until-writable-then-delivers"`.
2. Set `req.SendMessage = "writable-wait-probe"`.
3. Set `req.ExecTimeout = 25s` to allow busy-then-idle transition.

```go
import "time"

func Setup(t *testing.T, req *Request) error {
	req.Action = "default-waits-until-writable-then-delivers"
	req.SendMessage = "writable-wait-probe"
	req.ExecTimeout = 25 * time.Second
	return nil
}
```