# Scenario

**Feature**: SLACK_CHANNEL env fallback

```
SLACK_CHANNEL=C0... slack-msg send --token TOK MESSAGE -> OK
```

## Steps

1. No `--channel`; set `SLACK_CHANNEL` in env.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send", "--token", slackTestToken, "env channel msg"}
	req.Env = []string{"SLACK_CHANNEL=C0ALE44K5J6"}
	return nil
}
```
