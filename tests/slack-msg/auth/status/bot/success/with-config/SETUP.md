# Scenario

**Feature**: bot status with --config prints absolute config path

```
slack-msg auth status --config PATH -> Using config from: <abs> -> auth.test ok
```

## Steps

1. Materialize valid-config.json; args `auth status` then insert `--config`.
2. Token from config botToken (no CLI --token).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"auth", "status"}
	if err := withConfigArg(t, d, req, "valid-config.json", false); err != nil {
		return err
	}
	return nil
}
```
