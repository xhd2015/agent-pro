# Scenario

**Feature**: RenderSudoersLine includes wildcard args pattern for sing-box

```
# sing-box + "run -c *" -> line includes arg pattern
```

## Preconditions

- Rule has command and `ArgsPattern`.

## Steps

1. Set VPN sing-box rule with `run -c *` pattern.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/sudosetup"
)

func Setup(t *testing.T, req *Request) error {
	req.Action = "command_with_wildcard_args"
	req.Rule = sudosetup.Rule{
		Command:     "/opt/homebrew/bin/sing-box",
		ArgsPattern: "run -c *",
	}
	return nil
}
```