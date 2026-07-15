# Scenario

**Feature**: skill install bundles utility scripts

```
agent-pro skill --install verify-on-behalf-of-user --force <dir>
-> SKILL.md + scripts/enter-sandbox.sh
```

## Steps

1. Set install target under temp dir.
2. Run `agent-pro skill --install verify-on-behalf-of-user --force <target>`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	target := filepath.Join(req.TempDir, "installed-skill")
	req.Args = []string{"skill", "--install", "verify-on-behalf-of-user", "--force", target}
	req.Env = append(req.Env, "VERIFY_INSTALL_TARGET="+target)
	return nil
}
```