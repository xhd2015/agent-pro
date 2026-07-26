## Expected

- Registry file exists under `stub-tty-registry/`.
- `sessions/stub-tty/<agent-id>/tty.json` exists with `alive: true`.

```go
import (
	"encoding/json"
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("stub-tty run exit %d stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	assertFileExists(t, resp.RegistryPath)
	assertFileExists(t, resp.TTYJSONPath)
	data, err := os.ReadFile(resp.TTYJSONPath)
	if err != nil { t.Fatal(err) }
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil { t.Fatal(err) }
	if obj["alive"] != true { t.Fatalf("tty.json alive: got %v", obj["alive"]) }
	if obj["runner_id"] != "stub-tty" { t.Fatalf("runner_id: got %v", obj["runner_id"]) }
}
```
