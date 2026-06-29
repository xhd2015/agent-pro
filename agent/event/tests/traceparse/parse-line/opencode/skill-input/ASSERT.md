## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/opencode/skill-input`.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	if !resp.OK { t.Fatal("expected parse ok") }
	assertContains(t, resp.Output, `"tool_name":"Skill"`)
	assertContains(t, resp.Output, "git-fetch")
	assertContains(t, resp.Output, "skill install --general-agents")

}
```
