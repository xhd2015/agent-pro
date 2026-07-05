## Expected Output

Startup stderr should include the listen URL. Neither web process stdout nor stderr should contain human-readable agent event lines.

## Expected

- Web process **stderr** contains `listening at` (startup).
- Web process **stdout** does not contain `💬` or `[done]`.
- Web process **stderr** does not contain `💬` or `[done]`.

## Side Effects

- Session `events.jsonl` under `AGENT_RUN_HOME` still records assistant messages (API/UI), only the server terminal stays silent.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	stderr := webProcessStderrText(req)
	stdout := webProcessStdoutText(req)
	assert.Output(t, stderr, `<contains>
listening at
</contains>`)
	for _, stream := range []struct {
		name string
		text string
	}{
		{"stdout", stdout},
		{"stderr", stderr},
	} {
		if strings.Contains(stream.text, "💬") {
			t.Fatalf("web process %s must not contain agent message prefix 💬:\n%s", stream.name, stream.text)
		}
		if strings.Contains(stream.text, "[done]") {
			t.Fatalf("web process %s must not contain [done] agent print:\n%s", stream.name, stream.text)
		}
	}
}
```