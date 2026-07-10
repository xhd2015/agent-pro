# Scenario

**Feature**: Using config from: (none) without --config

```
Caller -> slack-msg send --token --channel MESSAGE -> (none) line on stdout
```

## Steps

1. Inherit config-none setup.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send", "--token", slackTestToken, "--channel", "C0ALE44K5J6", "Hello"}
	return nil
}
```
