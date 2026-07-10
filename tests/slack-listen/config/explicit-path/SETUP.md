# Scenario

**Feature**: startup logs absolute config path

```
slack-listen listen --config PATH -> Using config from: <absolute-path>
```

## Steps

1. Materialize valid-config fixture and pass `--config`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if err := withConfigArg(t, req, "valid-config.json", false); err != nil {
		return err
	}
	req.Daemon = true
	req.WantAgentCalls = 0
	req.BotToken = ""
	req.AppToken = ""
	return nil
}
```