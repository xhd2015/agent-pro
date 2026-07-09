# Scenario

**Bug**: `tty-watch run bash -c 'echo yes'` hangs forever when registry `.lock` is held

```
test holds flock(registry/.lock) EX
  -> tty-watch run bash -c 'echo yes'
  -> must return error promptly (not hang forever on acquireRegistryLock)
```

Observed on shared `~/.tty-watch` when another long-lived process (e.g. agent tooling)
keeps the registry flock: `tty-watch run bash -c "echo yes"` produces no output and
blocks until killed. `acquireRegistryLock` uses blocking `LOCK_EX` with no timeout.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-while-registry-lock-held"
	req.RunCommand = []string{"bash", "-c", "echo yes"}
	return nil
}
```
