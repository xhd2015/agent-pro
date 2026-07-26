# Scenario

**Feature**: top-level help lists commands and help topics; `--topic` only with `--help`

```
Caller -> slack-msg -h|--help -> usage + Help topics (add-missing-scope) -> exit 0
Caller -> slack-msg --help --topic add-missing-scope -> topic guideline -> exit 0
Caller -> slack-msg --topic … alone / unknown topic -> stderr + exit 1
```

## Preconditions

- No tokens or config required for help path.
- No `slack-msg help` subcommand; topic index lives only in root help.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h`, `--help`, and/or `--topic …` in `req.Args`.

## Context

- Root help must list subcommands `send`, `history`, `listen`, `channels`, `auth`, and `session`.
- Root help must list Help topics including `add-missing-scope` and usage line
  `slack-msg --help [--topic TOPIC]`.
- Soft/hard missing_scope stderr points operators at
  `slack-msg --help --topic add-missing-scope`.

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
