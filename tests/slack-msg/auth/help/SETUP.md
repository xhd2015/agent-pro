# Scenario

**Feature**: auth command help lists status without API calls

```
Caller -> slack-msg auth -h|--help -> usage lists status -> exit 0
```

## Preconditions

- No token required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` after `auth`.

## Context

- Help must list `status` subcommand.

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
