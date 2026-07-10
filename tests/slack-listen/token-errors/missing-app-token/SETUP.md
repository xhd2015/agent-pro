# Scenario

**Feature**: missing app token

```
Caller -> slack-listen listen --bot-token ... -> app token required
```

## Steps

1. Provide bot token only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.BotToken = slackTestBotToken
	return nil
}
```