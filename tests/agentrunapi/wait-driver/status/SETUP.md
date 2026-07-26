# Scenario

**Feature**: pure tty status parse / ready check

```
StatusStdout -> ParseTTYStatus -> screen, sendable
             -> IsSessionReadyFromStatus -> ready?
```

## Steps

1. Set mode `status`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "status"
	return nil
}
```
