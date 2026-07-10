# Scenario

**Feature**: single positional carries multi-word message

```
slack-msg send --token --channel CH "Hello from script" -> one MESSAGE arg
```

## Steps

1. One multi-word positional.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"Hello from script",
	}
	return nil
}
```
