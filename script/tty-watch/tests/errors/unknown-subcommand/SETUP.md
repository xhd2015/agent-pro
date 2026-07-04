# Scenario

**Feature**: unknown subcommand returns error

```
# bogus dispatch target
tty-watch not-a-real-subcommand -> error
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "error-cmd"
	return nil
}
```