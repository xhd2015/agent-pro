# Scenario

**Feature**: list reports observer (`tty-watch watch`) client count in WATCH column

```
# detached sleep + observer WS client held open
harness -> detached sleep -> dial observer -> tty-watch list -> WATCH=1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "list-table-watch"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```