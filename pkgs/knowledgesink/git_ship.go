package knowledgesink

import (
	"bytes"
	"fmt"
	"os"
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
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), gitErrMsg(stderr, err))
	}
	return nil
}

func gitOut(opts Opts, dir string, args ...string) (string, error) {
	stdout, stderr, err := gitRun(opts, dir, args...)
	if err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), gitErrMsg(stderr, err))
	}
	return strings.TrimSpace(stdout), nil
}

// gitErrMsg prefers actionable stderr (e.g. missing git-lfs) over success chatter.
func gitErrMsg(stderr string, err error) string {
	msg := strings.TrimSpace(stderr)
	if msg == "" && err != nil {
		return err.Error()
	}
	lines := strings.Split(msg, "\n")
	var keep []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		// git checkout prints these on stderr even when hooks later fail the command.
		if strings.HasPrefix(trim, "Switched to a new branch") ||
			strings.HasPrefix(trim, "Switched to and reset branch") ||
			strings.HasPrefix(trim, "Already on ") ||
			strings.HasPrefix(trim, "Reset branch ") {
			continue
		}
		keep = append(keep, trim)
	}
	if len(keep) > 0 {
		return strings.Join(keep, "\n")
	}
	if msg != "" {
		return msg
	}
	if err != nil {
		return err.Error()
	}
	return "unknown git error"
}

// withShipGit disables hooks for automated ship ops (global LFS post-checkout/commit
// must not fail headless sink when git-lfs is absent from PATH).
func withShipGit(opts Opts) (Opts, func(), error) {
	hooksDir, err := os.MkdirTemp("", "knowledgesink-no-hooks-*")
	if err != nil {
		return opts, func() {}, fmt.Errorf("mkdir hooks stub: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(hooksDir) }
	base := opts.GitFn
	if base == nil {
		base = defaultGitRunner
	}
	ship := opts
	ship.GitFn = func(dir string, args ...string) (string, string, error) {
		prefixed := append([]string{"-c", "core.hooksPath=" + hooksDir}, args...)
		return base(dir, prefixed...)
	}
	return ship, cleanup, nil
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

// ShipToMR commits listed paths on the current hub branch (must match origin/<target>),
// pushes HEAD:<feature> with merge_request options, and optionally ff-pushes to origin/<target>.
// It does not checkout a feature branch or master in the hub worktree.
func ShipToMR(opts Opts, hubDir string, ship *ShipResult, autoMerge bool) (*ShipGitResult, error) {
	if ship == nil {
		return nil, fmt.Errorf("nil ship result")
	}
	shipOpts, cleanup, err := withShipGit(opts)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if err := gitOK(shipOpts, hubDir, "rev-parse", "--is-inside-work-tree"); err != nil {
		return nil, fmt.Errorf("hub is not a git repo")
	}

	branch := strings.TrimSpace(ship.GitBranchName)
	msg := strings.TrimSpace(ship.GitCommitMsg)
	if branch == "" {
		return nil, fmt.Errorf("git_branch_name empty")
	}
	if msg == "" {
		return nil, fmt.Errorf("git_commit_msg empty")
	}

	verboseNotice(opts, "ship: resolve origin base")
	originRef, targetBranch, err := resolveOriginBase(shipOpts, hubDir)
	if err != nil {
		return nil, err
	}
	verboseNotice(opts, "ship: origin=%s target=%s remote_branch=%s", originRef, targetBranch, branch)

	headFull, err := gitOut(shipOpts, hubDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	originFull, err := gitOut(shipOpts, hubDir, "rev-parse", originRef)
	if err != nil {
		return nil, err
	}
	if headFull != originFull {
		headShort, _ := gitOut(shipOpts, hubDir, "rev-parse", "--short", "HEAD")
		originShort, _ := gitOut(shipOpts, hubDir, "rev-parse", "--short", originRef)
		if headShort == "" {
			headShort = headFull
		}
		if originShort == "" {
			originShort = originFull
		}
		return nil, fmt.Errorf("ship: hub HEAD must match %s (got %s, %s=%s); refusing to commit on a diverged workspace branch",
			originRef, headShort, originRef, originShort)
	}

	curBranch, _ := gitOut(shipOpts, hubDir, "rev-parse", "--abbrev-ref", "HEAD")
	verboseNotice(opts, "ship: commit on %s (no checkout)", curBranch)

	paths := ship.GitCommitFiles.AllPaths()
	if len(paths) == 0 {
		return nil, fmt.Errorf("git_commit_files empty")
	}

	addArgs := append([]string{"add", "--"}, paths...)
	verboseNotice(opts, "ship: git add -- %s", strings.Join(paths, ", "))
	if err := gitOK(shipOpts, hubDir, addArgs...); err != nil {
		return nil, err
	}
	verboseNotice(opts, "ship: git commit -m %q", msg)
	if err := gitOK(shipOpts, hubDir, "commit", "-m", msg); err != nil {
		return nil, err
	}
	commitFull, err := gitOut(shipOpts, hubDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	commit, err := gitOut(shipOpts, hubDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, err
	}
	verboseNotice(opts, "ship: committed %s on %s", commit, curBranch)

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
	stdout, stderr, perr := gitRun(shipOpts, hubDir, pushArgs...)
	combined := stdout + "\n" + stderr
	out := &ShipGitResult{Branch: branch, Commit: commit}
	if perr != nil {
		verboseNotice(opts, "ship: MR push options rejected; fallback plain push")
		stdout2, stderr2, perr2 := gitRun(shipOpts, hubDir, "push", "origin", "HEAD:"+branch)
		if perr2 != nil {
			return nil, fmt.Errorf("git push: %s", gitErrMsg(stderr2, perr2))
		}
		out.Warning = "merge_request push options rejected; pushed branch and emitted create-MR URL only"
		out.MRURL = extractOrBuildMRURL(shipOpts, hubDir, branch, stdout2+"\n"+stderr2)
		out.PushOption = false
	} else {
		out.PushOption = true
		out.MRURL = extractOrBuildMRURL(shipOpts, hubDir, branch, combined)
	}
	if out.MRURL != "" {
		verboseNotice(opts, "ship: mr %s", out.MRURL)
	} else {
		verboseNotice(opts, "ship: pushed %s (no MR URL parsed)", branch)
	}

	if autoMerge {
		verboseNotice(opts, "ship: auto-merge ff-only %s → origin/%s", commit, targetBranch)
		if err := autoMergeFFPush(shipOpts, hubDir, commitFull, targetBranch); err != nil {
			return out, fmt.Errorf("auto-merge: %w", err)
		}
		tip, _ := gitOut(shipOpts, hubDir, "rev-parse", "--short", "origin/"+targetBranch)
		out.Merged = true
		out.MergedAt = tip
		verboseNotice(opts, "ship: merged origin/%s @ %s", targetBranch, tip)
	}
	verboseNotice(opts, "ship: done branch=%s commit=%s", out.Branch, out.Commit)
	return out, nil
}

// autoMergeFFPush fast-forwards origin/<target> to commitSHA without checking out target locally.
func autoMergeFFPush(opts Opts, hubDir, commitSHA, targetBranch string) error {
	if err := gitOK(opts, hubDir, "fetch", "origin", targetBranch); err != nil {
		return err
	}
	originTarget := "origin/" + targetBranch
	if err := gitOK(opts, hubDir, "merge-base", "--is-ancestor", originTarget, commitSHA); err != nil {
		return fmt.Errorf("not fast-forward")
	}
	refspec := commitSHA + ":refs/heads/" + targetBranch
	if err := gitOK(opts, hubDir, "push", "origin", refspec); err != nil {
		return fmt.Errorf("push origin/%s: %w", targetBranch, err)
	}
	_, _ = gitOut(opts, hubDir, "fetch", "origin", targetBranch)
	return nil
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
