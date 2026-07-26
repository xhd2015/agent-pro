# Scenario

**Feature**: knownChannels map fast path with --config

```
slack-msg send --config cfg --channel #general -> knownChannels lookup -> C0ALE44K5J6
```

## Steps

1. Load valid-config fixture; pass `--config` + `--channel #general`.

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
	req.Args = append(req.Args, "--channel", "#general", "known map")
	return nil
}
```
