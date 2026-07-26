# Scenario

**Feature**: CLI --channel overrides config default

```
slack-msg send --config cfg --channel C0OTHERCHAN MESSAGE -> CLI wins
```

## Steps

1. valid-config with default C0ALE44K5J6; override to C0OTHERCHAN from slacktest list.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send"}
	if err := withConfigArg(t, d, req, "valid-config.json", false); err != nil {
		return err
	}
	req.Args = append(req.Args, "--channel", "C0OTHERCHAN", "override channel")
	return nil
}
```
