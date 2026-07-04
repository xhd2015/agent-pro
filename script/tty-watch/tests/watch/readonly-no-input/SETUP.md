# Scenario

**Feature**: watch does not forward stdin to the remote session

```
# cat session would echo stdin; watch must not send harness probe input
harness -> detached cat -> watch + stdin probe -> no SHOULD_NOT_ECHO in output
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-readonly"
	return nil
}
```