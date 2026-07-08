# Scenario

**Feature**: knownSkills includes loop-workflow

```
agent-pro skill loop-workflow show -> name: loop-workflow in output
```

## Steps

1. Invoke `agent-pro skill loop-workflow show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "loop-workflow", "show"}
	return nil
}
```