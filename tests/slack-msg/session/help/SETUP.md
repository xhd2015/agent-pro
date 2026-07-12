# Scenario

**Feature**: session top-level help lists reply and history

```
slack-msg session -h|--help -> lists reply / history; exit 0
```

## Steps

1. Leaf sets help flag after `session`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```
