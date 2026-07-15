# Scenario

**Feature**: agent-pro skills lists verify-on-behalf-of-user

```
agent-pro skills -> line with verify-on-behalf-of-user and description
```

## Steps

1. Invoke `agent-pro skills`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skills"}
	return nil
}
```