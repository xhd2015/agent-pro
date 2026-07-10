# Scenario

**Feature**: `send -h` shows help

```
Caller -> slack-msg send -h -> usage stdout -> exit 0
```

## Steps

1. Args `["send", "-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send", "-h"}
	return nil
}
```
