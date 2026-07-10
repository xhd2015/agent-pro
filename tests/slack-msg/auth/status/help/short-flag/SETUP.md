# Scenario

**Feature**: `auth status -h` shows help

```
Caller -> slack-msg auth status -h -> usage stdout -> exit 0
```

## Steps

1. Args `["auth", "status", "-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"auth", "status", "-h"}
	return nil
}
```
