# Scenario

**Feature**: attach resize updates session PTY dimensions (last-wins)

```
# wide-line session; attach resize cols=100; snapshot shows single unwrapped line
harness -> run wide-line sh -> detach -> attach resize -> snapshot not 80-col wrapped
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-resize"
	return nil
}
```