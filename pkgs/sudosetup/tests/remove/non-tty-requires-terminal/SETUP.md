# Scenario

**Feature**: Remove errors when drop-in exists but stdin is not a TTY

```
# installed + !stdin TTY -> error before sudo rm
```

## Preconditions

- Drop-in and manifest seeded.
- Stdin is not an interactive terminal.

## Steps

1. Seed installed state with `StdinIsTerminal = false`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "remove_non_tty"
	req.StdinIsTerminal = false
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("testuser", req.Rule.Command, "")
	return nil
}
```