# Scenario

**Feature**: send does not append carriage return or newline after message

```
# detached cat capture; send hello
harness -> detached cat capture -> tty-watch send session-N hello -> capture.bin == hello (len 5)
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "send-no-suffix"
	req.SendMessage = "hello"
	return nil
}
```