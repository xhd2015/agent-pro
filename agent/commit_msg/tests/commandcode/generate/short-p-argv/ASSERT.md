## Expected Output

Stdout includes the formatted title and description from the argv recorder JSON:

```text
<contains>
feat: short-p argv
</contains>
```

## Expected
- gen-commit-msg succeeds with commandcode argv recorder (mock JSON on stdout).
- stdout contains title `feat: short-p argv` and description `prompt via temp file`.
- Recorder argv capture file exists and includes a `-p` flag.
- The value immediately after `-p` is **short** (length ≤ `MaxCommandCodePromptArgBytes` = 2048).
- That `-p` value does **not** embed the full staged unified diff:
  - must not contain the substring `diff --git`
  - must not contain a multi-line `Git diff:` body (no `Git diff:` followed by a newline then more prompt text that looks like the old inline prompt)
- The short `-p` instruction points at file-based delivery: either the `-p` value is an existing path that was copied, or the text mentions reading a file / includes an absolute path token (case-insensitive `read` and/or a `/…` path).
- When the recorder copied a prompt file during the agent run (`CommandCodePromptCopyPath` non-empty content), that copy **does** contain `diff --git` and `Git diff:` (full prompt body lived in the file, not argv).
- HEAD subject is unchanged (no `--commit`).
- `--max-turns` `16` still present on argv (regression lock with unit test).

## Side Effects
- No new git commit.
- Argv capture written under `req.CommandCodeArgvPath`.
- Optional prompt file copy under `req.CommandCodePromptCopyPath` when production uses a readable path during the agent process.

## Exit Code
- Zero.

## Errors
- None expected from Run / gen-commit-msg once generate works; assertion failures are RED until file-based `-p` delivery is implemented (today production still puts the full `commitPrompt` on `-p`).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("gen-commit-msg commandcode short-p generate should succeed, got: %v\nstdout:\n%s\nstderr:\n%s",
			resp.Err, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "feat: short-p argv") {
		t.Fatalf("stdout missing title, got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "prompt via temp file") {
		t.Fatalf("stdout missing description, got:\n%s", resp.Stdout)
	}
	gotHEAD := GitHEADSubjectCmd(t, req.GitDir)
	if gotHEAD != req.HEADSubjectBefore {
		t.Fatalf("HEAD subject changed without --commit: before=%q after=%q",
			req.HEADSubjectBefore, gotHEAD)
	}

	if req.CommandCodeArgvPath == "" {
		t.Fatal("CommandCodeArgvPath not set")
	}
	raw, err := os.ReadFile(req.CommandCodeArgvPath)
	if err != nil {
		t.Fatalf("read argv capture %s: %v", req.CommandCodeArgvPath, err)
	}
	args := ParseNULSeparatedArgs(raw)
	if len(args) == 0 {
		t.Fatalf("empty argv capture at %s", req.CommandCodeArgvPath)
	}

	pVal := ArgvValueAfter(args, "-p")
	if pVal == "" {
		t.Fatalf("argv missing -p value; got %q", args)
	}

	// RED until implementer: production currently embeds full commitPrompt on -p.
	if len(pVal) > MaxCommandCodePromptArgBytes {
		t.Fatalf("-p value too long for file-based delivery: len=%d max=%d (full diff must not sit on argv)\n-p prefix: %q",
			len(pVal), MaxCommandCodePromptArgBytes, truncateRunes(pVal, 200))
	}
	if strings.Contains(pVal, "diff --git") {
		t.Fatalf("-p must not embed staged unified diff (found %q); use temp prompt file + short -p instruction\n-p len=%d prefix: %q",
			"diff --git", len(pVal), truncateRunes(pVal, 200))
	}
	// Old inline prompt starts with "Generate a brief git commit message" and includes "Git diff:\n".
	if strings.Contains(pVal, "Git diff:") {
		t.Fatalf("-p must not embed the full commit prompt body (found %q); deliver via temp file\n-p len=%d",
			"Git diff:", len(pVal))
	}

	lower := strings.ToLower(pVal)
	hasRead := strings.Contains(lower, "read")
	hasAbsPath := false
	for _, tok := range strings.Fields(pVal) {
		if strings.HasPrefix(tok, string(os.PathSeparator)) || filepath.IsAbs(tok) {
			hasAbsPath = true
			break
		}
	}
	// Also accept -p being exactly a prompt file path.
	if _, err := os.Stat(pVal); err == nil {
		hasAbsPath = true
	}
	if !hasRead && !hasAbsPath {
		t.Fatalf("-p short instruction should mention reading a file and/or include an absolute path; got %q", pVal)
	}

	foundMaxTurns := false
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--max-turns" && args[i+1] == "16" {
			foundMaxTurns = true
			break
		}
	}
	if !foundMaxTurns {
		t.Fatalf("argv missing --max-turns 16; got %q", args)
	}

	// If recorder copied the prompt file during the agent run, it must hold the full body.
	if req.CommandCodePromptCopyPath != "" {
		if st, err := os.Stat(req.CommandCodePromptCopyPath); err == nil && st.Size() > 0 {
			body, err := os.ReadFile(req.CommandCodePromptCopyPath)
			if err != nil {
				t.Fatalf("read prompt copy: %v", err)
			}
			text := string(body)
			if !strings.Contains(text, "diff --git") {
				t.Fatalf("prompt file copy missing %q (full staged diff should live in the file)\nlen=%d prefix: %q",
					"diff --git", len(text), truncateRunes(text, 200))
			}
			if !strings.Contains(text, "Git diff:") {
				t.Fatalf("prompt file copy missing %q\nlen=%d", "Git diff:", len(text))
			}
			if !strings.Contains(text, "LARGE_STAGED_DIFF_MARKER") {
				t.Fatalf("prompt file copy missing staged content marker LARGE_STAGED_DIFF_MARKER")
			}
		}
		// Empty copy is allowed only when -p itself was a readable path that was not
		// still present at recorder time — but then hasAbsPath/read must still hold.
		// Prefer soft check: if production passes path-only -p that exists at run time,
		// copy should be non-empty after GREEN implementation.
		if st, err := os.Stat(req.CommandCodePromptCopyPath); err == nil && st.Size() == 0 {
			// Require that implementer leaves a recoverable path in -p so the agent can read it.
			// Soft: do not fail solely on empty copy if -p clearly instructs file read with abs path.
			if !hasAbsPath {
				t.Fatalf("prompt file was not captured and -p has no absolute path token; agent cannot load full prompt\n-p=%q", pVal)
			}
		}
	}
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= max {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}
```
