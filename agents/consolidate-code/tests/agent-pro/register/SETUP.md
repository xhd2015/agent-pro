# Scenario

**Feature**: knownSkills includes consolidate-code

```
agent-pro skill consolidate-code --show -> name: consolidate-code in output
```

## Steps

1. Invoke `agent-pro skill consolidate-code --show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "consolidate-code", "--show"}
	return nil
}
```