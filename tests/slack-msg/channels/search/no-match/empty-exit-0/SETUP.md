# Scenario

**Feature**: no-match search returns empty JSON and exit 0

```
slack-msg channels search --json zzz-no-hit -> {"channels":[]} -> exit 0
```

## Steps

1. Token, `--json`, non-matching QUERY.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"channels", "search",
		"--token", slackTestToken,
		"--json",
		"zzz-no-hit",
	}
	return nil
}
```
