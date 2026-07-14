# Scenario

**Feature**: knownSkills includes run-the-loop

```
agent-pro skill run-the-loop show -> name: run-the-loop in output
```

## Steps

1. Invoke `agent-pro skill run-the-loop show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "run-the-loop", "show"}
	return nil
}
```