# Scenario

**Feature**: attach on unknown session id fails

```
# missing registry entry
tty-watch attach session-99999 -> exit 1; not found error
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-unknown"
	req.AttachID = "session-99999"
	return nil
}
```