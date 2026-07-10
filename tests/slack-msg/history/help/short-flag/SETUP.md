# Scenario

**Feature**: `history -h` shows help

```
Caller -> slack-msg history -h -> usage stdout -> exit 0
```

## Steps

1. Args `["history", "-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"history", "-h"}
	return nil
}
```
