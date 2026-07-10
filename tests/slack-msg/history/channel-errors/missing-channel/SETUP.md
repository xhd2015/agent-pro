# Scenario

**Feature**: missing channel for history

```
Caller -> slack-msg history --token TOK -> channel required
```

## Steps

1. Token only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"history", "--token", slackTestToken}
	return nil
}
```
