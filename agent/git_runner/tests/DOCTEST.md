# git_runner index write retry tests

Doc-style tests for `github.com/xhd2015/agent-pro/agent/git_runner` covering
transient git index error classification and `CommitWithRetry` recovery after
`index.lock` / `unable to write new index file` failures.

# DSN (Domain Specific Notion)

Production `gen-commit-msg --commit` (and any caller of `CommitWithRetry`) can
hit concurrent git writers that leave a stale **index.lock** or block rewriting
the git **index**. The original production symptom is:

```text
fatal: unable to write new index file
```

Participants and behaviors:

- **Caller** — e.g. `commit_msg` with `--commit` — asks `CommitWithRetry(dir, msg, attempts, noVerify)` to create a commit.
- **IsTransientIndexError** — classifies git stdout/stderr (+ optional `error`) as retryable when it mentions `index.lock`, `unable to write new index file`, or `unable to write index file`; rejects permanent failures (empty message, hook exit).
- **RemoveStaleIndexLock** — before each attempt, deletes leftover `index.lock` if present.
- **CommitWithRetry** — loops: clear lock → `git commit -m …` → on transient error back off and retry; on non-transient error return immediately; on success return commit output.
- **Interference** — test harness injects stale lock, macOS `chflags uchg` on the index (then clears), or a failing pre-commit hook so recovery vs fail-fast is observable.

```
# classify path
doctest -> IsTransientIndexError(output) -> true|false

# recovery path
interference (stale lock | uchg-then-clear | hook fail)
  -> CommitWithRetry
  -> git commit attempt(s)
doctest <- commit subject / error / first-fail classification
```

## Version

0.0.2

## Decision Tree

```
agent/git_runner/tests/
├── DOCTEST.md
├── SETUP.md
├── classify/                              # Mode=classify — pure string classifier
│   ├── transient/                         # WantTransient=true
│   │   ├── unable-to-write-new-index-file/  exact production message
│   │   ├── index-lock-file-exists/          classic index.lock File exists
│   │   └── unable-to-write-index-file/      casing/variant without "new"
│   └── non-transient/                     # WantTransient=false
│       ├── empty-commit-message/
│       └── hook-failure/
└── commit-with-retry/                     # Mode=commit-with-retry — real temp repo
    ├── stale-index-lock/                  # leftover index.lock → commit succeeds
    ├── unable-to-write-then-clear/        # Darwin uchg force write fail, clear, retry (skip non-darwin)
    └── non-transient-hook-failure/        # failing pre-commit → no infinite retry
```

Parameter ranking (most → least significant):

1. **Mode** — classify strings vs exercise `CommitWithRetry` in a real repo
2. **Outcome class** (classify) — transient vs non-transient
3. **Error / interference shape** — production write message, lock, hook, empty message, uchg
4. **Platform** — Darwin-only uchg injection (leaf skips elsewhere)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `classify/transient/unable-to-write-new-index-file` | Exact `fatal: unable to write new index file` → transient |
| 2 | `classify/transient/index-lock-file-exists` | `Unable to create …/index.lock: File exists` → transient |
| 3 | `classify/transient/unable-to-write-index-file` | `error: Unable to write index file` variant → transient |
| 4 | `classify/non-transient/empty-commit-message` | Empty commit message abort → not transient |
| 5 | `classify/non-transient/hook-failure` | husky/pre-commit hook failure → not transient |
| 6 | `commit-with-retry/stale-index-lock` | Stale `index.lock` present → `CommitWithRetry` succeeds; subject matches |
| 7 | `commit-with-retry/unable-to-write-then-clear` | First commit under uchg fails with production message (transient); after clear, retry commits (Darwin) |
| 8 | `commit-with-retry/non-transient-hook-failure` | Failing pre-commit → `CommitWithRetry` fails without treating as transient |

## How to Run

```sh
doctest vet ./agent/git_runner/tests
doctest test -v ./agent/git_runner/tests
doctest test -v ./agent/git_runner/tests/classify
doctest test -v ./agent/git_runner/tests/commit-with-retry/stale-index-lock
doctest test -v ./agent/git_runner/tests/commit-with-retry/unable-to-write-then-clear
```

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/git_runner"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	// Mode is "classify" or "commit-with-retry".
	Mode string

	// classify
	ClassifyOutput string
	WantTransient  bool

	// commit-with-retry
	RepoDir      string
	Message      string
	MaxAttempts  int
	NoVerify     bool
	Interference string // "stale-lock" | "uchg-then-clear" | "hook-fail"
}

type Response struct {
	// classify
	Transient bool

	// commit-with-retry
	Output             string
	CommitErr          error
	Subject            string
	FirstFailOutput    string
	FirstFailErr       error
	FirstFailTransient bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	switch req.Mode {
	case "classify":
		got := git_runner.IsTransientIndexError(req.ClassifyOutput, nil)
		return &Response{Transient: got}, nil
	case "commit-with-retry":
		return runCommitWithRetryScenario(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runCommitWithRetryScenario(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.RepoDir == "" {
		return nil, fmt.Errorf("commit-with-retry requires RepoDir")
	}
	if req.Message == "" {
		req.Message = "feat: doctest commit"
	}
	if req.MaxAttempts < 1 {
		req.MaxAttempts = 5
	}

	resp := &Response{}

	switch req.Interference {
	case "stale-lock":
		lockPath, err := git_runner.IndexLockPath(req.RepoDir)
		if err != nil {
			return nil, fmt.Errorf("IndexLockPath: %w", err)
		}
		if err := os.WriteFile(lockPath, []byte("stale"), 0644); err != nil {
			return nil, fmt.Errorf("write stale index.lock: %w", err)
		}
	case "uchg-then-clear":
		if runtime.GOOS != "darwin" {
			t.Skip("unable-to-write-then-clear requires Darwin chflags uchg")
		}
		indexPath, err := gitIndexPath(req.RepoDir)
		if err != nil {
			return nil, err
		}
		if err := exec.Command("chflags", "uchg", indexPath).Run(); err != nil {
			return nil, fmt.Errorf("chflags uchg: %w", err)
		}
		failOut, failErr := git_runner.Commit("should-fail-once", false).Dir(req.RepoDir).Run()
		resp.FirstFailOutput = string(failOut)
		resp.FirstFailErr = failErr
		resp.FirstFailTransient = git_runner.IsTransientIndexError(string(failOut), failErr)
		if clearErr := exec.Command("chflags", "nouchg", indexPath).Run(); clearErr != nil {
			return resp, fmt.Errorf("chflags nouchg: %w", clearErr)
		}
		time.Sleep(20 * time.Millisecond)
	case "hook-fail":
		// repo already prepared with failing pre-commit by leaf Setup
	case "":
		// no interference
	default:
		return nil, fmt.Errorf("unknown Interference %q", req.Interference)
	}

	out, err := git_runner.CommitWithRetry(req.RepoDir, req.Message, req.MaxAttempts, req.NoVerify)
	resp.Output = string(out)
	resp.CommitErr = err
	if err == nil {
		subj, logErr := git_runner.NewCommand("log", "-1", "--format=%s").Dir(req.RepoDir).Output()
		if logErr != nil {
			return resp, fmt.Errorf("git log after commit: %w", logErr)
		}
		resp.Subject = strings.TrimSpace(string(subj))
	}
	return resp, nil
}

func gitIndexPath(dir string) (string, error) {
	gdirOut, err := git_runner.NewCommand("rev-parse", "--absolute-git-dir").Dir(dir).Output()
	if err != nil {
		return "", fmt.Errorf("absolute-git-dir: %w", err)
	}
	return filepath.Join(strings.TrimSpace(string(gdirOut)), "index"), nil
}
```
