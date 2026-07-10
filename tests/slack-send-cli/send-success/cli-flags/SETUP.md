# Scenario

**Feature**: send with CLI --token and --channel

```
slack-send --token --channel MESSAGE -> slacktest -> OK
```

## Steps

1. Args via flags; message positional.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--token", slackTestToken,
		"--channel", "#general",
		"Hello world",
	}
	return nil
}
```