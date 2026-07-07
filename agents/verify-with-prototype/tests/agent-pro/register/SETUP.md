# Scenario

**Feature**: knownSkills includes verify-with-prototype

```
agent-pro skill verify-with-prototype show -> name: verify-with-prototype in output
```

## Steps

1. Invoke `agent-pro skill verify-with-prototype show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "verify-with-prototype", "show"}
	return nil
}
```