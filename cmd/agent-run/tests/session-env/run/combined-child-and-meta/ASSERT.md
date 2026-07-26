---
label: e2e
---

## Expected

- Exit code 0.
- Env probe: `PATH` starts with absolute tools dir; `FOO=bar`; `GROK_HOME=<config-home>`.
- Env probe does **not** require `AGENT_RUNNER_CONFIG_HOME=` (must not be the sole
  signal for config home).
- `meta.json` for the session has:
  - `prepend_paths` containing the abs tools path
  - `env` containing `"FOO=bar"`
  - `agent_runner_config_home` equal to abs config home

## Exit Code

0

```go
import (
	"strings"
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
	if strings.Contains(probe, "AGENT_RUNNER_CONFIG_HOME=") {
		// Allowed only if GROK_HOME is also set; config home must surface as GROK_HOME.
		// Soft note: presence is not required; if present, still require GROK_HOME above.
	}

	meta := readMetaJSON(t, req.Home, req.SessionID)
	assertMetaStringSliceEquals(t, meta, "prepend_paths", []string{req.PrependPathDir})
	assertMetaStringSliceEquals(t, meta, "env", []string{"FOO=bar"})
	if got := metaString(meta, "agent_runner_config_home"); got != req.AgentRunnerConfigHome {
		t.Fatalf("meta.agent_runner_config_home=%q want %q; meta=%v", got, req.AgentRunnerConfigHome, meta)
	}
}
```
