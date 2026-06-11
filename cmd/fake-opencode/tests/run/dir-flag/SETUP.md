## Preconditions
- The command passes `--dir`.

## Steps
1. Run fake opencode with an explicit working directory argument.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","stdout_events":[]}`)
    req.Args = []string{"run", "--format", "json", "--dir", filepath.Join(req.TempDir, "work"), "--mock-config", req.MockConfigPath, "hello"}
    return nil
}
```

