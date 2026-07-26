# Scenario

**Feature**: public channel ID used as-is

```
slack-msg send --channel C0ALE44K5J6 -> no API list lookup -> send OK
```

## Steps

1. Pass channel ID directly.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"direct C",
	}
	return nil
}
```
