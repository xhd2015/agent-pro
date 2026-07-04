# Scenario

**Feature**: list reports attacher (`tty-watch attach`) client count in ATTACHED column

```
# detached sleep + attach WS client held open
harness -> detached sleep -> dial attach -> tty-watch list -> ATTACHED=1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "list-table-attach"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```