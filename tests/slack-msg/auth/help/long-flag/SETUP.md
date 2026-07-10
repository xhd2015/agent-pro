# Scenario

**Feature**: `auth --help` shows help

```
Caller -> slack-msg auth --help -> usage stdout -> exit 0
```

## Steps

1. Args `["auth", "--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"auth", "--help"}
	return nil
}
```
