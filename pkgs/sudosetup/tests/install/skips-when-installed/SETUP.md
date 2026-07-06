# Scenario

**Feature**: EnsureInstalled skips visudo/install when already installed

```
# IsInstalled=true -> EnsureInstalled returns without visudo/install
```

## Preconditions

- Drop-in and manifest already match.

## Steps

1. Seed fully installed state.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "skips_when_installed"
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("testuser", req.Rule.Command, "")
	return nil
}
```