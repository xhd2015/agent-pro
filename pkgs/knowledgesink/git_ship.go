package knowledgesink

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GitRunner runs git -C dir args. Nil → exec.Command.
type GitRunner func(dir string, args ...string) (stdout, stderr string, err error)

func defaultGitRunner(dir string, args ...string) (string, string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func gitRun(opts Opts, dir string, args ...string) (string, string, error) {
	fn := opts.GitFn
	if fn == nil {
		fn = defaultGitRunner
	}
	return fn(dir, args...)
}

func gitOK(opts Opts, dir string, args ...string) error {
	_, stderr, err := gitRun(opts, dir, args...)
	if err != nil {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return nil
}

func gitOut(opts Opts, dir string, args ...string) (string, error) {
	stdout, stderr, err := gitRun(opts, dir, args...)
	if err != nil {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(stdout), nil
}

// HubGitUser returns stripped username from git config user.email in hub.
func HubGitUser(opts Opts, hubDir string) (string, error) {
	email, err := gitOut(opts, hubDir, "config", "user.email")
	if err != nil {
		return "", fmt.Errorf("git user.email: %w", err)
	}
	user := UserFromEmail(email)
	if user == "" {
		return "", fmt.Errorf("git user.email empty")
	}
	return user, nil
}

// resolveOriginBase returns (originRef, targetBranchName) e.g. ("origin/master","master").
func resolveOriginBase(opts Opts, hubDir string) (originRef, target string, err error) {
	if e := gitOK(opts, hubDir, "fetch", "origin", "master"); e == nil {
		if _, e2 := gitOut(opts, hubDir, "rev-parse", "--verify", "origin/master"); e2 == nil {
			return "origin/master", "master", nil
		}
	}
	if e := gitOK(opts, hubDir, "fetch", "origin", "main"); e == nil {
		if _, e2 := gitOut(opts, hubDir, "rev-parse", "--verify", "origin/main"); e2 == nil {
			return "origin/main", "main", nil
		}
	}
	return "", "", fmt.Errorf("origin/master (or origin/main) not found")
}

// ShipGitResult is the outcome of host-side commit/push/(auto-merge).
type ShipGitResult struct {
	Branch     string
	Commit     string
	MRURL      string
	Merged     bool
	MergedAt   string // short tip of origin/<target> after auto-merge
	PushOption bool
	Warning    string
}

// ShipToMR stashes agent edits, branches from origin/master, commits, pushes with
// merge_request push options, optionally ff-merges into origin/master.
func ShipToMR(opts Opts, hubDir string, ship *ShipResult, autoMerge bool) (*ShipGitResult, error) {
	if ship == nil {
		return nil, fmt.Errorf("nil ship result")
	}
	if err := gitOK(opts, hubDir, "rev-parse", "--is-inside-work-tree"); err != nil {
		return nil, fmt.Errorf("hub is not a git repo")
	}

	branch := strings.TrimSpace(ship.GitBranchName)
	msg := strings.TrimSpace(ship.GitCommitMsg)
	verboseNotice(opts, "ship: resolve origin base")
	originRef, targetBranch, err := resolveOriginBase(opts, hubDir)
	if err != nil {
		return nil, err
	}
	verboseNotice(opts, "ship: origin=%s target=%s branch=%s", originRef, targetBranch, branch)

	prevBranch, _ := gitOut(opts, hubDir, "rev-parse", "--abbrev-ref", "HEAD")

	paths := ship.GitCommitFiles.AllPaths()
	if len(paths) == 0 {
		return nil, fmt.Errorf("git_commit_files empty")
	}
	verboseNotice(opts, "ship: stash %d path(s): %s", len(paths), strings.Join(paths, ", "))
	stashArgs := append([]string{"stash", "push", "-u", "-m", "knowledgesink-create-mr", "--"}, paths...)
	if err := gitOK(opts, hubDir, stashArgs...); err != nil {
		return nil, fmt.Errorf("stash agent files: %w", err)
	}
	stashed := true
	defer func() {
		if stashed {
			_, _, _ = gitRun(opts, hubDir, "stash", "pop")
		}
		_ = restoreCheckout(opts, hubDir, prevBranch)
	}()

	_, _, _ = gitRun(opts, hubDir, "branch", "-D", branch)
	verboseNotice(opts, "ship: checkout -B %s %s", branch, originRef)
	if err := gitOK(opts, hubDir, "checkout", "-B", branch, originRef); err != nil {
		return nil, fmt.Errorf("checkout -B %s %s: %w", branch, originRef, err)
	}
	verboseNotice(opts, "ship: stash pop onto %s", branch)
	if err := gitOK(opts, hubDir, "stash", "pop"); err != nil {
		return nil, fmt.Errorf("restore agent files onto branch: %w", err)
	}
	stashed = false

	addArgs := append([]string{"add", "--"}, paths...)
	verboseNotice(opts, "ship: git add -- %s", strings.Join(paths, " "))
	if err := gitOK(opts, hubDir, addArgs...); err != nil {
		return nil, err
	}
	verboseNotice(opts, "ship: git commit -m %q", msg)
	if err := gitOK(opts, hubDir, "commit", "-m", msg); err != nil {
		return nil, err
	}
	commit, err := gitOut(opts, hubDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, err
	}
	verboseNotice(opts, "ship: committed %s on %s", commit, branch)

	// Unrelated dirty files (or agent paths omitted from result.json) may remain;
	// only the listed paths were added/committed.

	title := FormatMRTitle(opts.Source, msg)
	pushArgs := []string{
		"push", "origin", "HEAD:" + branch,
		"-o", "merge_request.create",
		"-o", "merge_request.target=" + targetBranch,
		"-o", "merge_request.title=" + title,
	}
	verboseNotice(opts, "ship: git push origin HEAD:%s (merge_request.create → %s)", branch, targetBranch)
	stdout, stderr, perr := gitRun(opts, hubDir, pushArgs...)
	combined := stdout + "\n" + stderr
	out := &ShipGitResult{Branch: branch, Commit: commit}
	if perr != nil {
		verboseNotice(opts, "ship: MR push options rejected; fallback plain push")
		stdout2, stderr2, perr2 := gitRun(opts, hubDir, "push", "origin", "HEAD:"+branch)
		if perr2 != nil {
			msg := strings.TrimSpace(stderr2)
			if msg == "" {
				msg = perr2.Error()
			}
			return nil, fmt.Errorf("git push: %s", msg)
		}
		out.Warning = "merge_request push options rejected; pushed branch and emitted create-MR URL only"
		out.MRURL = extractOrBuildMRURL(opts, hubDir, branch, stdout2+"\n"+stderr2)
		out.PushOption = false
	} else {
		out.PushOption = true
		out.MRURL = extractOrBuildMRURL(opts, hubDir, branch, combined)
	}
	if out.MRURL != "" {
		verboseNotice(opts, "ship: mr %s", out.MRURL)
	} else {
		verboseNotice(opts, "ship: pushed %s (no MR URL parsed)", branch)
	}

	if autoMerge {
		verboseNotice(opts, "ship: auto-merge ff-only %s → origin/%s", branch, targetBranch)
		if err := autoMergeToMaster(opts, hubDir, branch, targetBranch); err != nil {
			return out, fmt.Errorf("auto-merge: %w", err)
		}
		tip, _ := gitOut(opts, hubDir, "rev-parse", "--short", "origin/"+targetBranch)
		out.Merged = true
		out.MergedAt = tip
		verboseNotice(opts, "ship: merged origin/%s @ %s", targetBranch, tip)
	}
	verboseNotice(opts, "ship: done branch=%s commit=%s", out.Branch, out.Commit)
	return out, nil
}

func autoMergeToMaster(opts Opts, hubDir, featureBranch, targetBranch string) error {
	if err := gitOK(opts, hubDir, "fetch", "origin", targetBranch); err != nil {
		return err
	}
	if err := gitOK(opts, hubDir, "checkout", "-B", targetBranch, "origin/"+targetBranch); err != nil {
		return fmt.Errorf("checkout %s: %w", targetBranch, err)
	}
	if err := gitOK(opts, hubDir, "merge", "--ff-only", featureBranch); err != nil {
		return fmt.Errorf("not fast-forward")
	}
	if err := gitOK(opts, hubDir, "push", "origin", targetBranch); err != nil {
		return fmt.Errorf("push origin/%s: %w", targetBranch, err)
	}
	_, _ = gitOut(opts, hubDir, "fetch", "origin", targetBranch)
	return nil
}

func restoreCheckout(opts Opts, hubDir, branch string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" || branch == "HEAD" {
		return nil
	}
	return gitOK(opts, hubDir, "checkout", branch)
}

func extractOrBuildMRURL(opts Opts, hubDir, branch, pushOutput string) string {
	if u := extractMRURLFromPush(pushOutput); u != "" {
		return u
	}
	return buildNewMRURL(opts, hubDir, branch)
}

func extractMRURLFromPush(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "https://"); i >= 0 {
			u := strings.TrimSpace(line[i:])
			u = strings.TrimRight(u, ".")
			if strings.Contains(u, "merge_request") {
				return strings.Fields(u)[0]
			}
		}
	}
	return ""
}

func buildNewMRURL(opts Opts, hubDir, branch string) string {
	origin, err := gitOut(opts, hubDir, "remote", "get-url", "origin")
	if err != nil || origin == "" {
		return ""
	}
	web := originToWebBase(origin)
	if web == "" {
		return ""
	}
	enc := strings.ReplaceAll(branch, "/", "%2F")
	return web + "/-/merge_requests/new?merge_request%5Bsource_branch%5D=" + enc
}

func originToWebBase(origin string) string {
	origin = strings.TrimSpace(origin)
	origin = strings.TrimSuffix(origin, ".git")
	if strings.HasPrefix(origin, "git@") {
		rest := strings.TrimPrefix(origin, "git@")
		host, path, ok := strings.Cut(rest, ":")
		if !ok {
			return ""
		}
		return "https://" + host + "/" + strings.TrimPrefix(path, "/")
	}
	if strings.HasPrefix(origin, "https://") || strings.HasPrefix(origin, "http://") {
		return strings.TrimSuffix(origin, "/")
	}
	if strings.HasPrefix(origin, "ssh://git@") {
		rest := strings.TrimPrefix(origin, "ssh://git@")
		host, path, ok := strings.Cut(rest, "/")
		if !ok {
			return ""
		}
		if h, _, cut := strings.Cut(host, ":"); cut {
			host = h
		}
		return "https://" + host + "/" + path
	}
	return ""
}
