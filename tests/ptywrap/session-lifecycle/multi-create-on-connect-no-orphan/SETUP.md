# Scenario

**Feature**: repeated create-on-connect + normal close must not orphan shells

Models the production leak: clients (e.g. LocalTerminal, StrictMode remounts)
open `WS /api/terminal?name=...` **without** `session_id` (creates a new shell
each time), then disconnect with close code **1000** (no DELETE, no 4000).

After N such cycles, **zero** shell processes from those sessions should still
be running. Today each cycle leaves a live `bash --rcfile ~/.ptywrap-bashrc -i`
and exhausts `kern.tty.ptmx_max`.

```
# N times
WS create-on-connect -> close 1000
# expected: RunningProcessCount == 0
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "lifecycle-multi-create-orphan"
	req.RepeatCount = 5
	return nil
}
```
