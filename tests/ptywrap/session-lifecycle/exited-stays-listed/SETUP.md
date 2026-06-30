# Scenario

**Feature**: exited sessions stay in list with status exited

```
# short-lived cmd
create true -> wait exit -> list shows status exited
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "lifecycle-exited"
	req.Command = []string{"true"}
	return nil
}
```