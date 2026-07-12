# Scenario

**Feature**: starting + sendable no is not ready

```
fixture(screen status: starting, sendable: no) -> IsSessionReady false
```

## Preconditions

- Not-ready fixture from harness helper.

## Steps

1. Set StatusStdout to not-ready fixture.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StatusStdout = statusNotReadyFixture()
	return nil
}
```
