# Scenario

**Feature**: default run — no args, default channel and text

```
slack-send -> defaultChannelId + "Hello slack" -> OK
```

## Steps

1. Empty `req.Args`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = nil
	return nil
}
```