# Scenario

**Feature**: default live send to configured default channel

```
slack-send (no args) -> real API -> OK
```

## Steps

1. Inherit integration grouping setup (repo config).
2. Explicit default invocation: no channel/text args.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = nil
	req.UseRepoConfig = true
	return nil
}
```