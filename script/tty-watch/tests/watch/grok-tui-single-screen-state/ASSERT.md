## Expected

- Watch capture contains `Grok Build Beta` exactly once (latest screen state only).
- Watch capture must not contain orphaned SGR fragments like `[38;2;` (useless chars in
  the input area when true-color sequences are partially stripped).
- No line longer than 120 characters of smeared braille/content from stacked redraws.
- Observer mode is readonly.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	count := strings.Count(combined, "Grok Build Beta")
	if count != 1 {
		t.Fatalf("watch stacked %d screen snapshots (want 1 latest state), got %q", count, combined)
	}
	if strings.Contains(combined, "[38;2;") {
		t.Fatalf("watch leaked orphaned true-color SGR fragments, got %q", combined)
	}
	for _, line := range strings.Split(combined, "\n") {
		if len(line) > 120 {
			t.Fatalf("watch smeared long line (%d chars) from stacked redraws: %q", len(line), line)
		}
	}
}
```