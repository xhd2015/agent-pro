# Scenario

**Feature**: knownSkills includes sound-fix

```
agent-pro skill sound-fix show -> name: sound-fix in output
```

## Steps

1. Invoke `agent-pro skill sound-fix show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "sound-fix", "show"}
	return nil
}
```