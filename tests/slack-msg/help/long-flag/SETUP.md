# Scenario

**Feature**: `--help` shows top-level help

```
Caller -> slack-msg --help -> usage stdout -> exit 0
```

## Steps

1. Args `["--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
