# Scenario

**Feature**: detect rejects manifest with wrong username

```
# drop-in + manifest present but manifest user != current user
Detect -> Installed=false
```

## Preconditions

- Drop-in and manifest both exist.
- Manifest username differs from `Config.Username`.

## Steps

1. Seed matching drop-in.
2. Seed manifest with `otheruser`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "manifest_mismatch_user"
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("otheruser", req.Rule.Command, "")
	return nil
}
```