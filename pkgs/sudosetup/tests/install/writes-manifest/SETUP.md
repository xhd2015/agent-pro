# Scenario

**Feature**: EnsureInstalled writes manifest JSON matching rule after install

```
# successful install -> sudo-setup-manifest.json with username/command/args_pattern
```

## Preconditions

- No prior install.
- Install path succeeds.

## Steps

1. Use VPN-style rule with args pattern for manifest fields.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/sudosetup"
)

func Setup(t *testing.T, req *Request) error {
	req.Action = "writes_manifest"
	req.Config.CacheDirName = "remote-agent"
	req.Config.SudoersName = "remote-agent-sing-box"
	req.Rule = sudosetup.Rule{
		Command:     "/opt/homebrew/bin/sing-box",
		ArgsPattern: "run -c *",
	}
	return nil
}
```