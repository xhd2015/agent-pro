# Scenario

**Feature**: `listen -h` shows help

```
Caller -> slack-msg listen -h -> usage stdout -> exit 0
```

## Steps

1. Args `["-h"]` (listen prepended by harness).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```
