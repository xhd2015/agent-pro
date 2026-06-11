## Expected
- Exit code 0.
- The FIFO contains a JSON line with `type` = `"question"` and the correct question text.

```go
import (
    "testing"
    "time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)

    fifoPath := ""
    for _, env := range req.Env {
        if strings.HasPrefix(env, "TEST_FIFO_PATH=") {
            fifoPath = env[len("TEST_FIFO_PATH="):]
            break
        }
    }
    fifoData := readFifo(t, fifoPath, 5*time.Second)
    lines := parseJSONLines(t, fifoData)
    if len(lines) == 0 {
        t.Fatal("no JSON lines in FIFO")
    }
    q := lines[0]
    if qType, _ := q["type"].(string); qType != "question" {
        t.Fatalf("type = %q, want %q", qType, "question")
    }
    if qID, _ := q["id"].(string); qID != "1" {
        t.Fatalf("id = %q, want %q", qID, "1")
    }
    if qText, _ := q["question"].(string); qText != "What is the target port?" {
        t.Fatalf("question = %q, want %q", qText, "What is the target port?")
    }
}
```
