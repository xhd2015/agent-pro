# Scenario

**Feature**: --config provides token and default channel

```
slack-send --config cfg MESSAGE -> defaults from JSON -> OK
```

## Steps

1. valid-config fixture; message only after `--config`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if err := withConfigArg(t, req, "valid-config.json", false); err != nil {
		return err
	}
	req.Args = append(req.Args, "Hello from config")
	return nil
}
```