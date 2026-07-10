# Scenario

**Feature**: root `--topic` only with `--help`; known topics print guidelines

```
# success (either flag order)
slack-msg --help --topic add-missing-scope
slack-msg --topic add-missing-scope --help
  -> topic body on stdout (trailing \n) -> exit 0

# errors
slack-msg --topic add-missing-scope
  -> stderr requires --help -> exit 1
slack-msg --help --topic <unknown>
  -> stderr unknown help topic: … -> exit 1
```

## Preconditions

- No tokens or config; no Slack API.
- Topic index is only in root help (no `--topic list`).

## Steps

1. Inherit clear Slack env from `help/SETUP.md`.
2. Ensure no mock API URL for pure help/topic paths.
3. Leaf sets `req.Args` for topic / help combination.

## Context

- Topic `add-missing-scope` is the operator guideline for missing OAuth scopes
  (e.g. `groups:read`); soft/hard scope messages point here.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SlackAPIURL = ""
	return nil
}
```
