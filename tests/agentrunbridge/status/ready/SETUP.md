# Scenario

**Feature**: banner + sendable yes is ready

```
fixture(screen status: banner, sendable: yes) -> IsSessionReady true
```

## Preconditions

- Fixture uses lowercase labels; values `banner` and `yes`.

## Steps

1. Set StatusStdout to ready fixture.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StatusStdout = statusReadyFixture()
	return nil
}
```
