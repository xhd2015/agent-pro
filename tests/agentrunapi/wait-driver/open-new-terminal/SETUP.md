# Scenario

**Feature**: OpenInNewTerminal with injectable OpenTerminal (no real iTerm)

```
OpenInNewTerminal(WorkspaceDir, FollowUp|FollowUpOpts, OpenTerminal)
  -> OpenTerminal(dir, followUp)
```

## Steps

1. Set mode `open_new_terminal`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "open_new_terminal"
	return nil
}
```
