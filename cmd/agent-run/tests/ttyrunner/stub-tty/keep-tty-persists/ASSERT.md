## Expected

- Registry persists; `tty.json` has `alive: true`.

```go
import (
	"encoding/json"
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertFileExists(t, resp.RegistryPath)
	data, err := os.ReadFile(resp.TTYJSONPath)
	if err != nil { t.Fatal(err) }
	var obj map[string]any
	_ = json.Unmarshal(data, &obj)
	if obj["alive"] != true { t.Fatalf("alive: got %v want true", obj["alive"]) }
}
```
