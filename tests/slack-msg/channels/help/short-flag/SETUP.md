# Scenario

**Feature**: `channels -h` shows help

```
Caller -> slack-msg channels -h -> usage stdout -> exit 0
```

## Steps

1. Args `["channels", "-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"channels", "-h"}
	return nil
}
```
