# Scenario

**Feature**: `general` normalized and resolved via API

```
slack-msg send --channel general -> normalize #general -> API list -> C0ALE44K5J6
```

## Steps

1. Channel name without `#` prefix.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "general",
		"resolve plain",
	}
	return nil
}
```
