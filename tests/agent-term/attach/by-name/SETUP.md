# Scenario

**Feature**: attach resolves session by renamed name

```
# rename then attach
create -> PATCH name -> WS probe by name succeeds
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "attach-by-name"
	req.RenameBeforeAttach = "attach-target"
	req.StartDaemon = true
	return nil
}
```