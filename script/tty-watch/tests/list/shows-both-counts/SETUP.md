# Scenario

**Feature**: list reports watch and attach counts concurrently

```
# detached sleep + observer + attach WS clients held open
harness -> detached sleep -> dial observer + attach -> tty-watch list -> WATCH=1 ATTACHED=1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "list-table-both"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```