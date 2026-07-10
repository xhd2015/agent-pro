# Scenario

**Feature**: `channels list -h` shows help

```
Caller -> slack-msg channels list -h -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "list", "-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"channels", "list", "-h"}
	return nil
}
```
