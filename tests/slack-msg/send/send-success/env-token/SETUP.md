# Scenario

**Feature**: SLACK_BOT_TOKEN env fallback

```
SLACK_BOT_TOKEN=... slack-msg send --channel CH MESSAGE -> OK
```

## Steps

1. No `--token`; set `SLACK_BOT_TOKEN` in env.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send", "--channel", "C0ALE44K5J6", "env token msg"}
	req.Env = []string{"SLACK_BOT_TOKEN=" + slackTestToken}
	return nil
}
```
