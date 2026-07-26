# Scenario

**Feature**: listen help flags print usage without Socket Mode connect

```
Caller -> slack-msg listen -h|--help -> usage on stdout -> exit 0 (no WS connect)
```

## Preconditions

- No tokens required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` in `req.Args`.

## Context

- Help must show `--token` (not `--bot-token`) and `slack-msg listen`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
