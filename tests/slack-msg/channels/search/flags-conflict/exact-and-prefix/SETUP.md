# Scenario

**Feature**: both --exact and --prefix rejected

```
Caller -> slack-msg channels search --token --exact --prefix gen -> exit 1
```

## Steps

1. Token, both flags, QUERY `gen`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"channels", "search",
		"--token", slackTestToken,
		"--exact",
		"--prefix",
		"gen",
	}
	return nil
}
```
