# Scenario

**Bug**: `tty-watch snapshot` replays screen frames at hardcoded 80×24 while the ptywrap
session was resized wider, wrapping long lines that should fit on one row.

```
# ?2026h + CUP draws 95-char marker on row 30; writer resize sets cols=100 rows=32
harness -> run (wide-line sh) -> detach -> writer WS resize -> snapshot <id>
```

Reproduces user report where status bars and long TUI lines wrap at 80 columns in
`snapshot` output even though `run` resized the session to ~140 columns.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "snapshot-session-dimensions-wide"
	return nil
}
```