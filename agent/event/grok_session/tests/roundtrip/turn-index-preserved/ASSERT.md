## Expected
- Every event's `turn_index` matches between events₁ and events₂.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Events1) != len(resp.Events2) {
		t.Fatalf("length mismatch: %d vs %d", len(resp.Events1), len(resp.Events2))
	}
	for i := range resp.Events1 {
		if grokTurnIndex(resp.Events1[i]) != grokTurnIndex(resp.Events2[i]) {
			t.Fatalf("events[%d] turn_index mismatch: %d vs %d", i,
				grokTurnIndex(resp.Events1[i]), grokTurnIndex(resp.Events2[i]))
		}
	}
}
```
