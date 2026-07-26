# Scenario

**Feature**: banner + sendable yes is ready

```
IsSessionReadyFromStatus(banner + yes) -> true
ParseTTYStatus -> screen=banner, sendable=yes
```

## Steps

1. Use ready fixture stdout.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StatusStdout = statusReadyFixture()
	return nil
}
```
