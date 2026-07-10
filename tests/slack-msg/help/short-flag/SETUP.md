# Scenario

**Feature**: `-h` shows top-level help

```
Caller -> slack-msg -h -> usage stdout -> exit 0
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
