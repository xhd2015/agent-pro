# Scenario

**Bug**: registry lock busy error is a short misleading one-liner without holders or process tree

```
# distinctive child holds exclusive flock on isolated registry/.lock
harness lockholder(PID, marker) EX on $TTY_WATCH_HOME/registry/.lock
  -> tty-watch run bash -c 'echo yes'
  -> must exit promptly with stderr diagnostics:
     registry lock busy + lock path + holders table + process tree
  -> not: hang forever, and not: only "another tty-watch run may be in progress"
```

When the registry exclusive flock is held, `tty-watch run` times out (~1.5s) and
must print rich **stderr** diagnostics so operators can identify the real holder
(often not another `tty-watch run`). Sibling leaf `bash-c-lock-held-no-hang`
only seals no-hang; this leaf locks the fixed richer error contract.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-while-registry-lock-held-diagnostics"
	req.RunCommand = []string{"bash", "-c", "echo yes"}
	return nil
}
```
