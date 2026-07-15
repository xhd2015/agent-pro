# Scenario

**Feature**: knownSkills includes summarize-a-skill

```
agent-pro skill summarize-a-skill --show -> name: summarize-a-skill in output
```

## Steps

1. Invoke `agent-pro skill summarize-a-skill --show`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "summarize-a-skill", "--show"}
	return nil
}
```
