## Preconditions
- A named FIFO pipe exists.
- `QUESTION_FIFO` env var points to it.
- The binary is invoked as `yield-pending-questions` with a JSON question arg.

## Steps
1. Create a FIFO pipe.
2. Set `QUESTION_FIFO` env var.
3. Invoke the binary with a question JSON argument.
4. Read the FIFO to verify the question was written.

```go
import (
    "syscall"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    fifo := createFifo(t, req.TempDir)
    req.Env = append(req.Env, "QUESTION_FIFO="+fifo, "TEST_FIFO_PATH="+fifo)
    req.Args = []string{`{"id":"1","question":"What is the target port?"}`}
    return nil
}
```
