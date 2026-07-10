# Scenario

**Feature**: config defaultChannelId as channel name

```
slack-msg send --config cfg MESSAGE -> resolve "#general" default -> C0ALE44K5J6
```

## Steps

1. Load default-channel-name fixture (defaultChannelId is `#general`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send"}
	if err := withConfigArg(t, req, "default-channel-name.json", false); err != nil {
		return err
	}
	req.Args = append(req.Args, "name default")
	return nil
}
```
