# Scenario

**Feature**: detect confirms non-interactive command works when installed

```
# installed + sudo -n <command> succeeds
Detect -> Installed=true, CanRunNonInteractive=true
```

## Preconditions

- Drop-in and manifest match current rule.
- Runner fakes successful `sudo -n <command>` probe.

## Steps

1. Seed installed state.
2. Enable `SudoNCommandOK` with sample output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "installed_and_runnable"
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("testuser", req.Rule.Command, "")
	req.SudoNTrueOK = true
	req.SudoNCommandOK = true
	req.SudoNCommandOutput = "hello\n"
	return nil
}
```