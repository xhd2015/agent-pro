# Scenario

**Feature**: ancestor grok present but no open-file session hit

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Procs = defaultAncestorChain()
	req.OpenFiles = map[int][]string{
		pidGrok: {"/tmp/unrelated.txt"},
	}
	req.PID = pidStart
	req.Args = []string{}
	return nil
}
```
