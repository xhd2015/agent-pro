# Scenario

**Feature**: missing channel

```
Caller -> slack-msg send --token TOK MESSAGE -> channel required
```

## Steps

1. Token + message only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send", "--token", slackTestToken, "Hello"}
	return nil
}
```
