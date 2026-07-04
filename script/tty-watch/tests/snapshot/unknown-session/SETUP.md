# Scenario

**Feature**: snapshot on unknown session id fails

```
# missing registry entry
tty-watch snapshot session-99999 -> error
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "snapshot-missing"
	req.SnapshotID = "session-99999"
	return nil
}
```