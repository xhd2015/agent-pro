# Scenario

**Feature**: session lifecycle after PTY exit

```
# exit retention
PTY exits -> status exited -> session remains in GET list until DELETE
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "lifecycle-exited"
	return nil
}
```