# Scenario

**Feature**: reject missing or multiple positional MESSAGE arguments

```
Caller -> slack-msg send [options] (wrong arg count) -> stderr message required / exactly one -> exit 1
```

## Preconditions

- Provide enough flags so failure is attributable to message validation only.

## Steps

1. Clear Slack env vars.
2. Set `--token` and `--channel` so token/channel checks pass first.
3. Leaf narrows message arg count (0 or 2+).

## Context

- Channel is never a positional arg in the send contract.

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
