# Scenario

**Feature**: EnsureInstalled skips TTY requirement when already installed

```
# installed + !stdin TTY -> noop (no visudo/install)
```

## Preconditions

- Drop-in and manifest match current rule.
- Stdin is not an interactive terminal.

## Steps

1. Seed installed state with `StdinIsTerminal = false`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "non_tty_skips_when_installed"
	req.StdinIsTerminal = false
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("testuser", req.Rule.Command, "")
	return nil
}
```