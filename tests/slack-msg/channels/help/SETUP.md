# Scenario

**Feature**: channels help flags print usage without API calls

```
Caller -> slack-msg channels -h|--help -> usage on stdout -> exit 0
```

## Preconditions

- No token required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` after `channels`.

## Context

- Help must list subcommands `list` and `search`.

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
