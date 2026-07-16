# Scenario

**Feature**: `sessions --help` documents list/print options including
`--grok-session-id`

```
agent-run sessions --help -> options + --grok-session-id
```

## Steps

1. Leaf runs `sessions --help` and asserts documented flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Args) == 0 {
		req.Args = []string{"sessions", "--help"}
	}
	return nil
}
```
