# Scenario

**Bug**: `tty-watch run bash -c 'echo yes'` leaves a KeepAlive `__serve__` zombie forever

```
harness PTY -> tty-watch run --session-id bash-c-no-orphan bash -c 'echo yes'
  -> parent prints yes + [Terminal exited] + exit 0
  -> __serve_bash_c_echo_yes__ must also exit (no orphan after ~2s grace)
```

Root cause: `run` always sets `KeepAlive: true`, so after the PTY child exits the
serve process blocks forever on `ctx.Done()` instead of shutting down. Parent
removes the registry entry, but the orphaned serve keeps running.

This also contributes to registry-lock / session-id races that make subsequent
`tty-watch run bash -c "echo yes"` hang forever on a shared `TTY_WATCH_HOME`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-bash-c-no-orphan-serve"
	req.CustomSessionID = "bash-c-no-orphan"
	req.RunCommand = []string{"bash", "-c", "echo yes"}
	return nil
}
```
