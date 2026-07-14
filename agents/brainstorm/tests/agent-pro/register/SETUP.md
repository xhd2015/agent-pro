# Scenario

**Feature**: knownSkills includes brainstorm

```
agent-pro skill brainstorm show -> name: brainstorm in output
```

## Steps

1. Invoke `agent-pro skill brainstorm show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "brainstorm", "show"}
	return nil
}
```