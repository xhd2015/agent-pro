# Scenario

**Feature**: `#general` resolved via conversations.list

```
slack-send --channel "#general" -> API list -> C0ALE44K5J6 -> send OK
```

## Steps

1. Args with channel name including `#`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--token", slackTestToken,
		"--channel", "#general",
		"resolve hash",
	}
	return nil
}
```