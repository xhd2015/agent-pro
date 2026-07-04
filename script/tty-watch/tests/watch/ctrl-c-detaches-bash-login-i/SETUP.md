# Scenario

**Feature**: User-reported working baseline — bash --login -i then watch Ctrl-C

```
# user flow that works: tty-watch run bash --login -i -> detach -> watch -> Ctrl-C
harness -> run bash --login -i -> detach -> watch -> inject \x03
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-ctrl-c-detaches-bash-login-i"
	return nil
}
```