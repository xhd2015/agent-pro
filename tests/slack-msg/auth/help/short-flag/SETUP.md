# Scenario

**Feature**: `auth -h` shows help

```
Caller -> slack-msg auth -h -> usage stdout -> exit 0
```

## Steps

1. Args `["auth", "-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"auth", "-h"}
	return nil
}
```
