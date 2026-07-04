# Scenario

**Feature**: list with empty registry produces no session lines

```
# fresh home, no sessions
tty-watch list -> empty registry -> no session-N lines
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "list-empty"
	return nil
}
```