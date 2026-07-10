# Scenario

**Feature**: bot status --json structured document

```
slack-msg auth status --token --json -> JSON status (no raw token)
```

## Steps

1. Flags `--token` and `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"auth", "status",
		"--token", slackTestToken,
		"--json",
	}
	return nil
}
```
