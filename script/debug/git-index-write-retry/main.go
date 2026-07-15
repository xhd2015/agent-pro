// git-index-write-retry — verify CommitWithRetry treats transient index
// write failures (including "unable to write new index file") as retryable
// and can complete a commit after interference.
//
// Usage (from repo root):
//
//	go run ./script/debug/git-index-write-retry
//
// Exit 0 = PASS (goal met). Non-zero = FAIL.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/git_runner"
)

func main() {
	root, err := gitRoot()
	if err != nil {
		fail("git root: %v", err)
	}
	outDir := filepath.Join(root, "script/debug/git-index-write-retry/out")
	_ = os.MkdirAll(outDir, 0o755)

	var fails []string
	var passes []string

	// --- 1) Classification: exact original error must be retryable ---
	exact := "fatal: unable to write new index file"
	if !git_runner.IsTransientIndexError(exact, nil) {
		fails = append(fails, `classifier does not treat "unable to write new index file" as transient`)
	} else {
		passes = append(passes, "classifier accepts exact index-write error")
	}

	related := "fatal: Unable to create '/tmp/repo/.git/index.lock': File exists."
	if !git_runner.IsTransientIndexError(related, nil) {
		fails = append(fails, "classifier does not treat index.lock contention as transient")
	} else {
		passes = append(passes, "classifier accepts index.lock contention")
	}

	if git_runner.IsTransientIndexError("fatal: empty commit message", nil) {
		fails = append(fails, "classifier wrongly treats empty commit message as transient")
	} else {
		passes = append(passes, "classifier rejects non-transient commit errors")
	}

	// --- 2) Stale lock recovery (CommitWithRetry) ---
	if err := scenarioStaleLock(outDir); err != nil {
		fails = append(fails, "stale lock recovery: "+err.Error())
	} else {
		passes = append(passes, "CommitWithRetry recovers from stale index.lock")
	}

	// --- 3) macOS: brief immutable index then clear + retry → commit succeeds ---
	if runtime.GOOS == "darwin" {
		if err := scenarioImmutableThenRetry(outDir); err != nil {
			fails = append(fails, "immutable-index retry: "+err.Error())
		} else {
			passes = append(passes, "CommitWithRetry recovers after transient index write failure (uchg)")
		}
	} else {
		passes = append(passes, "skip immutable-index scenario (non-darwin)")
	}

	// --- report ---
	fmt.Println("=== git-index-write-retry inspect ===")
	for _, p := range passes {
		fmt.Println("PASS:", p)
	}
	for _, f := range fails {
		fmt.Println("FAIL:", f)
	}
	summaryPath := filepath.Join(outDir, "summary.txt")
	_ = os.WriteFile(summaryPath, []byte(strings.Join(append(passes, fails...), "\n")+"\n"), 0o644)
	fmt.Println("out:", summaryPath)

	if len(fails) > 0 {
		fmt.Fprintf(os.Stderr, "\nFAIL: %d check(s) failed (goal not met)\n", len(fails))
		os.Exit(1)
	}
	fmt.Println("\nPASS: all checks — CommitWithRetry handles transient index write errors")
	os.Exit(0)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FAIL: "+format+"\n", args...)
	os.Exit(2)
}

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func initRepo(dir string) error {
	steps := [][]string{
		{"init", "--template=", "-b", "main"},
		{"config", "user.email", "inspect@example.com"},
		{"config", "user.name", "Inspect"},
		{"config", "core.hooksPath", "/dev/null"},
	}
	for _, args := range steps {
		if err := runGit(dir, args...); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		return err
	}
	if err := runGit(dir, "add", "seed.txt"); err != nil {
		return err
	}
	return runGit(dir, "commit", "-m", "seed")
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), string(out), err)
	}
	return nil
}

func scenarioStaleLock(outDir string) error {
	dir, err := os.MkdirTemp("", "idx-retry-stale-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := initRepo(dir); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
		return err
	}
	if err := runGit(dir, "add", "a.txt"); err != nil {
		return err
	}
	lockPath, err := git_runner.IndexLockPath(dir)
	if err != nil {
		return err
	}
	if err := os.WriteFile(lockPath, []byte("stale"), 0o644); err != nil {
		return err
	}
	out, err := git_runner.CommitWithRetry(dir, "feat: after stale lock", 5, false)
	_ = os.WriteFile(filepath.Join(outDir, "stale-lock.log"), out, 0o644)
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(out))
	}
	subj, err := git_runner.NewCommand("log", "-1", "--format=%s").Dir(dir).Output()
	if err != nil {
		return err
	}
	if got := strings.TrimSpace(string(subj)); got != "feat: after stale lock" {
		return fmt.Errorf("subject = %q", got)
	}
	return nil
}

// Simulate the original failure mode: index write fails once, then succeeds.
// On Darwin we use chflags uchg for one attempt, clear it, then CommitWithRetry
// should still succeed across attempts. We wrap a multi-attempt path by:
//  1. Setting uchg
//  2. Running commit once (expect fail with unable to write new index file)
//  3. Clearing uchg
//  4. CommitWithRetry (must succeed)
// Plus: prove classifier would have retried the failed attempt.
func scenarioImmutableThenRetry(outDir string) error {
	dir, err := os.MkdirTemp("", "idx-retry-uchg-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := initRepo(dir); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b\n"), 0o644); err != nil {
		return err
	}
	if err := runGit(dir, "add", "b.txt"); err != nil {
		return err
	}

	gdirOut, err := git_runner.NewCommand("rev-parse", "--absolute-git-dir").Dir(dir).Output()
	if err != nil {
		return err
	}
	indexPath := filepath.Join(strings.TrimSpace(string(gdirOut)), "index")

	// First attempt: force exact original error.
	if err := exec.Command("chflags", "uchg", indexPath).Run(); err != nil {
		return fmt.Errorf("chflags uchg: %w", err)
	}
	failOut, failErr := git_runner.Commit("should-fail-once", false).Dir(dir).Run()
	_ = os.WriteFile(filepath.Join(outDir, "uchg-first-attempt.log"), failOut, 0o644)
	if failErr == nil {
		_ = exec.Command("chflags", "nouchg", indexPath).Run()
		return fmt.Errorf("expected first commit to fail under uchg, succeeded: %s", string(failOut))
	}
	if !strings.Contains(string(failOut), "unable to write new index file") {
		_ = exec.Command("chflags", "nouchg", indexPath).Run()
		return fmt.Errorf("expected unable to write new index file, got: %s", string(failOut))
	}
	if !git_runner.IsTransientIndexError(string(failOut), failErr) {
		_ = exec.Command("chflags", "nouchg", indexPath).Run()
		return fmt.Errorf("first failure not classified as transient: %s", string(failOut))
	}

	// Interference clears (as concurrent process finishes / unlock).
	if err := exec.Command("chflags", "nouchg", indexPath).Run(); err != nil {
		return fmt.Errorf("chflags nouchg: %w", err)
	}
	// Small pause like retry backoff.
	time.Sleep(20 * time.Millisecond)

	out, err := git_runner.CommitWithRetry(dir, "feat: recovered after index write fail", 5, false)
	_ = os.WriteFile(filepath.Join(outDir, "uchg-retry.log"), out, 0o644)
	if err != nil {
		return fmt.Errorf("CommitWithRetry after clear: %v\n%s", err, string(out))
	}
	subj, err := git_runner.NewCommand("log", "-1", "--format=%s").Dir(dir).Output()
	if err != nil {
		return err
	}
	if got := strings.TrimSpace(string(subj)); got != "feat: recovered after index write fail" {
		return fmt.Errorf("subject = %q", got)
	}
	return nil
}
