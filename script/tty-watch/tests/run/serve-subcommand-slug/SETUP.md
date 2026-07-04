# Scenario

**Feature**: `tty-watch run` reexec child uses `__serve_{slug}__` token derived from command argv

```
tty-watch run sleep 300
  -> child argv[1] = __serve_sleep_300__ (not legacy __serve__)
  -> registry PID -> ps confirms slug token
```

## Steps

1. Set phase `run-registers` with default `sleep 300` command.
2. Harness detaches after registry appears; Response carries session id + registry state.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-registers"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```