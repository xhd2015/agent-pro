# Scenario

**Feature**: session --help

```
slack-msg session --help -> same usage as -h
```

## Steps

1. Args: session --help.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "--help"}
	return nil
}
```
