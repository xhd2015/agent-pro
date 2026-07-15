# Scenario

**Feature**: knownSkills includes followup

```
agent-pro skill followup --show -> name: followup in output
```

## Steps

1. Invoke `agent-pro skill followup --show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "followup", "--show"}
	return nil
}
```