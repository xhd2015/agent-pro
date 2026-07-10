# Scenario

**Feature**: missing bot token

```
Caller -> slack-listen listen --app-token ... -> bot token required
```

## Steps

1. Provide app token only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AppToken = slackTestAppToken
	return nil
}
```