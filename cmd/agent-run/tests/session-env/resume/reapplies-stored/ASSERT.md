---
label: e2e
---

## Expected

- Exit code 0.
- Child PATH starts with stored abs tools dir.
- Child has `FOO=bar`.
- Child has `GROK_HOME=<stored config home>` even without resume flag.
- Meta still holds the original three fields (unchanged or equivalent).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	probe := readEnvProbe(t, req.EnvProbePath)
	assertProbePATHPrefixed(t, probe, req.PrependPathDir)
	assertProbeHasKV(t, probe, "FOO", "bar")
	assertProbeHasKV(t, probe, "GROK_HOME", req.AgentRunnerConfigHome)

	meta := readMetaJSON(t, req.Home, req.SessionID)
	assertMetaStringSliceEquals(t, meta, "prepend_paths", []string{req.PrependPathDir})
	assertMetaStringSliceEquals(t, meta, "env", []string{"FOO=bar"})
	if got := metaString(meta, "agent_runner_config_home"); got != req.AgentRunnerConfigHome {
		t.Fatalf("meta.agent_runner_config_home=%q want %q", got, req.AgentRunnerConfigHome)
	}
}
```
