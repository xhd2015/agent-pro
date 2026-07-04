# Scenario

**Feature**: kill on unknown session id fails

```
# missing registry entry
tty-watch kill session-99999 -> error
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "kill-missing"
	req.KillID = "session-99999"
	return nil
}
```