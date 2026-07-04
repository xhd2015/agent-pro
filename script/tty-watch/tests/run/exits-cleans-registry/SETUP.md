# Scenario

**Feature**: attached command exit removes registry entry

```
# short-lived command exits while attached
harness PTY -> tty-watch true -> host exits -> registry empty
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-exit-clean"
	req.RunCommand = []string{"true"}
	return nil
}
```