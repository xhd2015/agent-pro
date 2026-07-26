# Scenario

**Branch**: agent binary integration with mock backend

```
mock server + opencode/pi -> Paris in output
```

## Preconditions
- These branches test real binary integration: the mock server runs as a backend and the
  actual agent binary (`opencode` or `pi`) is spawned to talk to it.
- Each leaf verifies the binary's stdout output and that the mock server's events file
  records the incoming LLM requests.
- Spawns must not use the developer's real agent home. The opencode leaf sets temporary
  `HOME` and `OPENCODE_CONFIG_DIR` so user plugins/hooks under `~/.config/opencode` are
  not loaded.

## Steps
1. The root `runBinary` creates a temp events file, passes `--events-file` to the mock server,
   spawns the external binary, then reads events from the file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // EventsFile is auto-created by runBinary if not set.
    // Each leaf sets BinaryCmd and BinaryEnv in its own Setup.
    t.Logf("integration Setup: BinaryCmd=%v", req.BinaryCmd)
    return nil
}
```
