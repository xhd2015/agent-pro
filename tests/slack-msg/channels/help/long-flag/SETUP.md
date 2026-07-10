# Scenario

**Feature**: `channels --help` shows help

```
Caller -> slack-msg channels --help -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"channels", "--help"}
	return nil
}
```
