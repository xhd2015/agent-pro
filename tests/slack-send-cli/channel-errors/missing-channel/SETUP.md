# Scenario

**Feature**: missing channel

```
Caller -> slack-send --token TOK MESSAGE -> channel required
```

## Steps

1. Inherit channel-errors setup.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--token", slackTestToken, "Hello"}
	return nil
}
```