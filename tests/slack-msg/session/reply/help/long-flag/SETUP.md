# Scenario

**Feature**: session reply --help

```
slack-msg session reply --help -> same as -h
```

## Steps

1. Args: session reply --help.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "reply", "--help"}
	return nil
}
```
