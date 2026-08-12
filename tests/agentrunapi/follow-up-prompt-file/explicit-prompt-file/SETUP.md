# Scenario

**Feature**: explicit `PromptFile` emits given path; no re-write / no argv body

```
BuildFollowUpCommand(
  PromptFile=/abs/given/path.txt, Prompt may be empty or ignored,
  PromptSpillDir=tmp, Open)
  -> --prompt-file=/abs/given/path.txt
  -> do not write under PromptSpillDir
  -> do not put Prompt after --
```

## Steps

1. Write given prompt file under `d.DOCTEST_CASE/given/path.txt`.
2. Set `PromptFile` to that absolute path.
3. Leave `Prompt` empty (or short ignored body).
4. Inject empty `PromptSpillDir` under `d.DOCTEST_CASE/spill` to prove no auto-spill.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "sess-explicit-pf"
	req.Open = true
	// Caller already spilled; Prompt ignored for argv delivery.
	req.Prompt = ""
	given, err := writeCaseFile(t, d, "given/path.txt", "pre-spilled-body\n")
	if err != nil {
		return err
	}
	req.PromptFile = given
	spill, err := ensureSpillDir(t, d, "spill")
	if err != nil {
		return err
	}
	req.PromptSpillDir = spill
	return nil
}
```
