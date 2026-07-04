# Scenario

**Feature**: attach forwards typed stdin to the PTY master

```
# detached cat; attach types unique marker; marker echoed on attach stdout
harness -> detached cat -> tty-watch attach <id> + stdin -> ATTACH_STDIN_MARKER visible
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-forwards-stdin"
	req.AttachInput = "ATTACH_STDIN_MARKER"
	return nil
}
```