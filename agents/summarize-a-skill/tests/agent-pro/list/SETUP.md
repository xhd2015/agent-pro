# Scenario

**Feature**: agent-pro skills lists summarize-a-skill with description

```
agent-pro skills -> Available skills listing includes summarize-a-skill + description
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
