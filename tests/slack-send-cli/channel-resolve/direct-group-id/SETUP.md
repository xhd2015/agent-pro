# Scenario

**Feature**: private group ID used as-is

```
slack-send --channel G024BE91L -> passthrough -> send OK
```

## Steps

1. Pass group ID directly.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--token", slackTestToken,
		"--channel", "G024BE91L",
		"direct G",
	}
	return nil
}
```