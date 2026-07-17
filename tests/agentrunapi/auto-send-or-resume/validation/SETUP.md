# Scenario

**Feature**: AutoSendOrResume input gates (no dispatch side effects)

```
empty SessionID | Open+Detach
  -> error
  -> RunCalls=SendCalls=ResumeCalls=0
```

## Preconditions

- Hooks installed so we can prove zero dispatch on validation failure.
- Store may be present but unused for these errors.

## Steps

1. Install hooks; leaves set invalid field combinations.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InstallHooks = true
	req.ExpectNoHooks = true
	return nil
}
```
