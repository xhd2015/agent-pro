# Scenario

**Feature**: agent-pro skills lists establish-a-loop with description

```
agent-pro skills -> Available skills listing includes establish-a-loop + description
```

## Steps

1. Invoke `agent-pro skills` with no arguments.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skills"}
	return nil
}
```