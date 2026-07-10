# Scenario

**Feature**: `channels list --help` shows help

```
Caller -> slack-msg channels list --help -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "list", "--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"channels", "list", "--help"}
	return nil
}
```
