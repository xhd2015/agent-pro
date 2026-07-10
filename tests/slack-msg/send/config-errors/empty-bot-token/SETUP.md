# Scenario

**Feature**: config with empty botToken and no CLI override

```
Caller -> slack-msg send --config empty.json MESSAGE -> botToken is empty in
```

## Steps

1. Materialize empty-token fixture; pass `--config` + message.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send"}
	if err := withConfigArg(t, req, "empty-token-config.json", false); err != nil {
		return err
	}
	req.Args = append(req.Args, "Hello")
	return nil
}
```
