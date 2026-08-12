# Scenario

**Feature**: run help surface documents `--prompt-file`

```
read pkgs/agentruncli/run_cmd.go (runHelp + flags)
  -> contains --prompt-file
```

## Steps

1. Mode `cli_help` (pure source; no Handle / no stdio swap).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "cli_help"
	return nil
}
```
