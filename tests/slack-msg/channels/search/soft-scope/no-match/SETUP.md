# Scenario

**Feature**: search empty after soft-skip private still exits 0 with warning

```
slack-msg channels search --json --token zzz-no-hit
  -> public listed but no name match; private skipped
  -> {"channels":[]} + stderr warning
  -> exit 0
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
