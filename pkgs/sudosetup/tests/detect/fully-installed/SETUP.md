# Scenario

**Feature**: detect reports installed when drop-in and manifest match

```
# matching drop-in + manifest for current user and rule
Detect -> Installed=true
```

## Preconditions

- Sudoers drop-in and manifest both match `Config.Username` and `Rule`.

## Steps

1. Seed drop-in line and manifest JSON for testuser + hello.sh.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "fully_installed"
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("testuser", req.Rule.Command, "")
	return nil
}
```