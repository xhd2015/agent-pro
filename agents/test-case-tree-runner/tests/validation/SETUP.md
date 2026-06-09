## Preconditions
- `test-case-tree-runner` is installed and available in PATH
- Run from this directory so that `InputDir` relative paths resolve against `testdata/`

## Steps
1. Set `req.InputDir` to the path of the fixture to test (done by each leaf SETUP.md)
2. Run `test-case-tree-runner compile <InputDir>`
3. Capture stdout/stderr and exit code

```go
import "os/exec"

type Request struct {
	InputDir string
}

type Response struct {
	Output string
	Passed bool
}

func Setup(t *testing.T, req *Request) error {
	_ = req
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	cmd := exec.Command("test-case-tree-runner", "compile", req.InputDir)
	out, _ := cmd.CombinedOutput()
	return &Response{
		Output: string(out),
		Passed: cmd.ProcessState.ExitCode() == 0,
	}, nil
}
```
