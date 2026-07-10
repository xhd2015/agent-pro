# Scenario

**Feature**: `-h` shows help

```
Caller -> slack-listen listen -h -> usage stdout -> exit 0
```

## Steps

1. Args `["-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```