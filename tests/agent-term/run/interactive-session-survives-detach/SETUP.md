# Scenario

**Feature**: Ctrl-C during interactive run detaches client; session keeps running

```
# detach without killing remote session
harness PTY -> agent-term run sleep 60 -> SIGINT -> session still running in list
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "run-detach-survives"
	req.StartDaemon = true
	return nil
}
```