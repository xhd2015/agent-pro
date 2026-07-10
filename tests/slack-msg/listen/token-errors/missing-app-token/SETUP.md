# Scenario

**Feature**: missing app token for listen

```
Caller -> slack-msg listen --token ... -> app token required
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
