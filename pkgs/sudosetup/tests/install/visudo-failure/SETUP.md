# Scenario

**Feature**: EnsureInstalled aborts when visudo validation fails

```
# visudo -cf returns error -> no manifest written
```

## Preconditions

- Not installed.
- Runner fakes visudo failure.

## Steps

1. Set `VisudoFails = true`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "visudo_failure"
	req.VisudoFails = true
	return nil
}
```