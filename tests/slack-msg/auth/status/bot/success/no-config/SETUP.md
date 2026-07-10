# Scenario

**Feature**: bot status without config prints Using config from: (none)

```
slack-msg auth status --token -> Using config from: (none) -> auth.test ok
```

## Steps

1. CLI `--token` only; no `--config`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"auth", "status",
		"--token", slackTestToken,
	}
	return nil
}
```
