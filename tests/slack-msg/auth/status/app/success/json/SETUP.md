# Scenario

**Feature**: app status --json structured document

```
slack-msg auth status --app --app-token --json -> JSON app status
```

## Steps

1. Flags `--app`, `--app-token`, `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"auth", "status",
		"--app",
		"--app-token", slackTestAppToken,
		"--json",
	}
	return nil
}
```
