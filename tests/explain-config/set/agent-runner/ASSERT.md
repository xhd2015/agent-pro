## Expected

- Exit 0.
- Stdout empty (or whitespace only).
- `config.json` contains `"agent_runner": "codex"`.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("expected empty stdout, got %q", resp.Stdout)
	}
	if resp.ConfigJSON == "" {
		t.Fatal("expected config.json to exist")
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(resp.ConfigJSON), &cfg); err != nil {
		t.Fatalf("parse config: %v (%s)", err, resp.ConfigJSON)
	}
	if cfg["agent_runner"] != "codex" {
		t.Fatalf("agent_runner = %#v, want codex (config=%s)", cfg["agent_runner"], resp.ConfigJSON)
	}
}
```
