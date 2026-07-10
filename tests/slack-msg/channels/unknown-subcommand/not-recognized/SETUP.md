# Scenario

**Feature**: unrecognized channels subcommand name

```
Caller -> slack-msg channels not-a-command -> stderr + exit 1
```

## Steps

1. Args `["channels", "not-a-command"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"channels", "not-a-command"}
	return nil
}
```
