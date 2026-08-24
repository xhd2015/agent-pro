## Expected

- No error.
- Output is valid JSON object with fields:
  - `session_id` = fixture id
  - `state` = `running`
  - `file_active` = true
  - `pid_checked` = true
  - `pids` = array length 1 with `pid` 5001, `name` `grok`, `cmd` containing grok
- No ANSI escape sequences in output.

## Errors

- None.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.Output == "" {
		t.Fatal("FormatStatusJSON output empty")
	}
	assertNoANSI(t, resp.Output)

	var doc struct {
		SessionID  string `json:"session_id"`
		State      string `json:"state"`
		FileActive bool   `json:"file_active"`
		PIDChecked bool   `json:"pid_checked"`
		Path       string `json:"path"`
		PIDs       []struct {
			PID  int    `json:"pid"`
			Name string `json:"name"`
			Cmd  string `json:"cmd"`
		} `json:"pids"`
	}
	if err := json.Unmarshal([]byte(resp.Output), &doc); err != nil {
		t.Fatalf("json.Unmarshal: %v\nraw=%s", err, resp.Output)
	}
	assertEqualString(t, "session_id", doc.SessionID, req.SessionID)
	assertEqualString(t, "state", doc.State, "running")
	assertEqualBool(t, "file_active", doc.FileActive, true)
	assertEqualBool(t, "pid_checked", doc.PIDChecked, true)
	if !strings.HasSuffix(doc.Path, "summary.json") {
		t.Fatalf("path = %q, want …/summary.json", doc.Path)
	}
	if strings.HasPrefix(doc.Path, "~") {
		t.Fatalf("JSON path must be absolute, got %q", doc.Path)
	}
	if len(doc.PIDs) != 1 {
		t.Fatalf("pids len = %d, want 1", len(doc.PIDs))
	}
	assertEqualInt(t, "pids[0].pid", doc.PIDs[0].PID, 5001)
	assertEqualString(t, "pids[0].name", doc.PIDs[0].Name, "grok")
	if !strings.Contains(doc.PIDs[0].Cmd, "grok") {
		t.Fatalf("pids[0].cmd = %q", doc.PIDs[0].Cmd)
	}
}
```
