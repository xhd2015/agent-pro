# Scenario

**Bug**: second `tty-watch run` after exiting the first shows empty `tty-watch list`

```
# user flow: tty-watch run codex -> exit -> tty-watch run codex -> list empty
# root cause: first __serve_{slug}__ zombie RemoveRegistry races with reused session-N id
first run true (fast exit) -> second run sleep 600 (live) -> wait grace -> list
```

Reproduces the reported `tty-watch run codex` then `tty-watch list` empty on the
second run. `true` stands in for a codex session that exits; `sleep 600` stands in
for the second long-lived codex session still running when list is queried.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "list-second-run-after-exit"
	return nil
}
```