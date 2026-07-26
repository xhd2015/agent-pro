# Scenario

**Feature**: listen startup logs absolute config path

```
slack-msg listen --config PATH -> Using config from: <absolute-path>
```

## Steps

1. Materialize valid-config fixture and pass `--config`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := withConfigArg(t, d, req, "valid-config.json", false); err != nil {
		return err
	}
	req.Daemon = true
	req.WantAgentCalls = 0
	req.BotToken = ""
	req.AppToken = ""
	return nil
}
```
