# Scenario

**Feature**: Remove deletes drop-in and manifest, flushes sudo cache

```
# installed -> sudo rm drop-in -> delete manifest -> sudo -k
```

## Preconditions

- Drop-in and manifest exist.

## Steps

1. Seed installed state.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "removes_drop_in_and_manifest"
	req.SeedSudoersLine = installedSeedLine("testuser", req.Rule.Command, "")
	req.SeedManifest = installedManifestSeed("testuser", req.Rule.Command, "")
	return nil
}
```