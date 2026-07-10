# Scenario

**Feature**: `channels search -h` shows help

```
Caller -> slack-msg channels search -h -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "search", "-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"channels", "search", "-h"}
	return nil
}
```
