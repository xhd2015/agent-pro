# Scenario

**Feature**: public channel ID used as-is

```
slack-msg send --channel C0ALE44K5J6 -> no API list lookup -> send OK
```

## Steps

1. Pass channel ID directly.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"direct C",
	}
	return nil
}
```
