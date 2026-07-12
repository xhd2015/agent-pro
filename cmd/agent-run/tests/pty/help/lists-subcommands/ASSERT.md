## Expected

- Exit code 0.
- Stdout lists `stats` and `kill-orphans` subcommands.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, ``+
		`<contains>
<start-with>
  stats
</start-with>
<start-with>
  kill-orphans
</start-with>
</contains>`)
	assertTrailingNewline(t, resp.Stdout, "pty --help stdout")
}
```
