# Scenario

**Feature**: detect rejects manifest with wrong command path

```
# drop-in + manifest present but manifest command != Rule.Command
Detect -> Installed=false
```

## Preconditions

- Drop-in and manifest both exist.
- Manifest command path differs from current rule.

## Steps

1. Seed drop-in for hello.sh.
2. Seed manifest with a different command path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "manifest_mismatch_command"
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("testuser", "/other/path/hello.sh", "")
	return nil
}
```