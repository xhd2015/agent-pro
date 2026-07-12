# Scenario

**Bug**: `run --open` leaves detached `__serve` reparented to PID 1; harness
must reclaim so TestGenerated cases do not leak PTYs

```
agent-run run --agent-runner grok-tty --open "pty-open-cleanup"
  + instant attach + hold TUI
  -> registry/serve present after open returns
  -> reclaimServesUnderHome(home)
  -> serve process gone
```

## Steps

1. Run open-cleanup flow (open + reclaim in `Run`).
2. Assert serve was alive before reclaim and dead after.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "open-cleanup"
	req.OpenInstantAttach = true
	req.Prompt = "pty-open-cleanup"
	req.GrokTTYCommand = fakeTUIHoldSeconds(30)
	req.Args = []string{
		"run", "--agent-runner", "grok-tty", "--open", req.Prompt,
	}
	req.ExecTimeout = 90 * time.Second
	return nil
}
```
