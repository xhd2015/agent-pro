# Scenario

**Feature**: top-level help lists send, history, and listen

```
Caller -> slack-msg -h|--help -> usage listing three commands -> exit 0
```

## Preconditions

- No tokens or config required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` in `req.Args`.

## Context

- Help must mention subcommands `send`, `history`, and `listen`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
