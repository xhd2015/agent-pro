# Scenario

**Feature**: `tty-watch list` prints an aligned table header when sessions exist

```
# detached sleep session, no clients
harness -> detached sleep -> tty-watch list -> SESSION UPTIME WATCH ATTACHED COMMAND header + aligned row
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "list-table-header"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```