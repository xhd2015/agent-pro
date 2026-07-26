# Scenario

**Feature**: Using config from: (none) without --config

```
Caller -> slack-msg send --token --channel MESSAGE -> (none) line on stdout
```

## Steps

1. Inherit config-none setup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "--token", slackTestToken, "--channel", "C0ALE44K5J6", "Hello"}
	return nil
}
```
