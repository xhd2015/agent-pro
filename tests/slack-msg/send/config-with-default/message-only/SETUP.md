# Scenario

**Feature**: --config provides token and default channel

```
slack-msg send --config cfg MESSAGE -> defaults from JSON -> OK
```

## Steps

1. valid-config fixture; message only after `--config`.

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
	req.Args = append(req.Args, "Hello from config")
	return nil
}
```
