package knowledgesink

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

// withShipGitIfHooksDisabled applies the hook suppression unless the caller asked to
// keep hooks. Wrapping the decision keeps every call site on one path.
func withShipGitIfHooksDisabled(opts Opts) (Opts, func(), error) {
	if opts.HonorHooks {
		return opts, func() {}, nil
	}
	return withShipGit(opts)
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

	// Tip is the commit both the feature branch and the target branch were
	// published from. After a merge it is the merge commit, not Commit: the
	// pre-merge commit is no longer a descendant of the target and cannot be
	// pushed to it.
	Tip string
	// Synced is true when origin/<target> moved into HEAD during this ship.
	Synced bool
	// Incoming is how many commits arrived from origin/<target>.
	Incoming int
	// Resolved lists conflicted paths the resolver cleared.
	Resolved []string
	// Swept lists the extra dirty paths this commit carried alongside the ones
	// result.json named, so the same knowledge does not sit uncommitted and block
	// the next KB update.
	Swept []string
	// Skipped lists dirty paths deliberately left alone (binary, nested repo,
	// scratch) with the reason, so "committed everything" is never a claim.
	Skipped []SweptPath
	// LeaseTip is the commit the feature branch pointed at before a retry, used to
	// move it by lease rather than by force.
	LeaseTip string
	// Attempts is how many publish rounds ran; 1 means no peer moved the target.
	Attempts int
	// Foreign is how many commits were already on HEAD and absent from
	// origin/<target> before this ship.
	Foreign int
}

// SweptPath is one dirty path this ship refused to commit.
type SweptPath struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// ShipToMR commits the paths named by result.json together with the other
// committable dirt in the hub, merges origin/<target> (resolving conflicts with
// the configured resolver), and publishes the resulting tip: first to the
// feature branch with merge_request options, then to origin/<target>.
//
// It does not checkout a feature branch or master in the hub worktree, and it
// never rebases: a merge has one stop, so a conflict is a bounded job.
func ShipToMR(opts Opts, hubDir string, ship *ShipResult, autoMerge bool) (*ShipGitResult, error) {
	if ship == nil {
		return nil, fmt.Errorf("nil ship result")
	}
	release, err := acquireShipLock(opts, hubDir)
	if err != nil {
		return nil, err
	}
	defer release()

	shipOpts, cleanup, err := withShipGitIfHooksDisabled(opts)
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
	paths := ship.GitCommitFiles.AllPaths()
	if len(paths) == 0 {
		return nil, fmt.Errorf("git_commit_files empty")
	}

	verboseNotice(opts, "ship: resolve origin base")
	originRef, targetBranch, err := resolveOriginBase(shipOpts, hubDir)
	if err != nil {
		return nil, err
	}
	verboseNotice(opts, "ship: origin=%s target=%s remote_branch=%s", originRef, targetBranch, branch)

	curBranch, _ := gitOut(shipOpts, hubDir, "rev-parse", "--abbrev-ref", "HEAD")
	verboseNotice(opts, "ship: commit on %s (no checkout)", curBranch)

	out := &ShipGitResult{Branch: branch}

	// Commits already on HEAD but absent from origin/<target> before this ship adds
	// its own. Counted before the commit so "foreign" means someone else's work, not
	// the sink's own commit reflected back at it.
	preCommitForeign := commitsAhead(shipOpts, hubDir, originRef)

	// 1. Commit first. git refuses to merge over local changes that the incoming
	// commit also touches, and it refuses before it detects a conflict, so merging
	// before committing would leave the resolver with nothing to resolve.
	swept, skipped := planSweep(shipOpts, hubDir, paths)
	all := append(append([]string(nil), paths...), swept...)
	verboseNotice(opts, "ship: git add -- %s", strings.Join(all, ", "))
	if err := gitOK(shipOpts, hubDir, append([]string{"add", "--"}, all...)...); err != nil {
		return out, err
	}
	verboseNotice(opts, "ship: git commit -m %q", msg)
	if err := gitOK(shipOpts, hubDir, "commit", "-m", msg); err != nil {
		return out, err
	}
	out.Swept = swept
	out.Skipped = skipped
	commitFull, err := gitOut(shipOpts, hubDir, "rev-parse", "HEAD")
	if err != nil {
		return out, err
	}
	out.Commit, _ = gitOut(shipOpts, hubDir, "rev-parse", "--short", "HEAD")
	verboseNotice(opts, "ship: committed %s on %s (swept=%d skipped=%d)",
		out.Commit, curBranch, len(swept), len(skipped))

	// 2. Merge origin/<target> into HEAD, resolving conflicts. The returned tip is
	// the merge commit when there is one; pushing anything earlier would be
	// rejected by the target branch.
	syncRes, err := syncBeforePublish(opts, hubDir, targetBranch)
	_ = preCommitForeign
	if err != nil {
		return out, err
	}
	if syncRes != nil {
		out.Synced = syncRes.Synced
		out.Incoming = syncRes.Incoming
		out.Resolved = syncRes.Resolved
		out.Foreign = preCommitForeign
		out.Tip = syncRes.Tip
	}
	if out.Tip == "" {
		out.Tip = commitFull
	}
	if out.Foreign > 0 {
		verboseNotice(opts, "ship: merging %d pre-existing local commit(s) into %s besides this sink commit",
			out.Foreign, targetBranch)
	}

	// 3/4. Publish the same tip to both refs, retrying as a unit. A peer that pushes
	// between our merge and our publish makes the target reject us; the fix is to
	// fetch, merge, and publish again -- not to push the pre-merge commit, which is
	// no longer a descendant of the target.
	maxAttempts := opts.PublishAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultPublishAttempts
	}
	title := FormatMRTitle(opts.Source, msg)
	// A commit id as the source needs the destination fully qualified, otherwise
	// git cannot tell whether it means a branch or a tag.
	featureRef := "refs/heads/" + branch
	for attempt := 1; ; attempt++ {
		out.Attempts = attempt
		if err := pushFeatureBranch(shipOpts, hubDir, out, branch, featureRef, targetBranch, title); err != nil {
			return out, err
		}
		if !autoMerge {
			verboseNotice(opts, "ship: done branch=%s commit=%s", out.Branch, out.Commit)
			return out, nil
		}
		verboseNotice(opts, "ship: ff origin/%s → %s", targetBranch, out.Tip)
		err := autoMergeFFPush(shipOpts, hubDir, out.Tip, targetBranch)
		if err == nil {
			tip, _ := gitOut(shipOpts, hubDir, "rev-parse", "--short", originRef)
			out.Merged = true
			out.MergedAt = tip
			verboseNotice(opts, "ship: merged origin/%s @ %s", targetBranch, tip)
			verboseNotice(opts, "ship: done branch=%s commit=%s tip=%s", out.Branch, out.Commit, out.Tip)
			return out, nil
		}
		if !isRetryablePush(err) || attempt >= maxAttempts {
			return out, fmt.Errorf("auto-merge: %w", err)
		}
		// Someone else published first. Re-merge, then publish the new tip; the
		// feature branch is this sink's own, so it is moved by lease rather than
		// left pointing at a commit the target will never contain.
		verboseNotice(opts, "ship: publish rejected (attempt %d/%d); re-merging", attempt, maxAttempts)
		previous := out.Tip
		resync, serr := syncBeforePublish(opts, hubDir, targetBranch)
		if serr != nil {
			return out, serr
		}
		if resync != nil {
			out.Synced = out.Synced || resync.Synced
			out.Incoming += resync.Incoming
			out.Resolved = append(out.Resolved, resync.Resolved...)
			out.Tip = resync.Tip
		}
		out.LeaseTip = previous
	}
}

// pushFeatureBranch publishes the tip to the sink's feature branch with the merge
// request options, falling back to a plain push when the server rejects them.
func pushFeatureBranch(opts Opts, hubDir string, out *ShipGitResult, branch, featureRef, targetBranch, title string) error {
	pushArgs := []string{
		"push", "origin", out.Tip + ":" + featureRef,
		"-o", "merge_request.create",
		"-o", "merge_request.target=" + targetBranch,
		"-o", "merge_request.title=" + title,
	}
	if out.LeaseTip != "" {
		// Only on a retry, and only ever against this sink's own branch: the lease
		// names the commit we published last round, so a third party's commit on the
		// same ref is still refused.
		pushArgs = append(pushArgs, "--force-with-lease="+featureRef+":"+out.LeaseTip)
	}
	verboseNotice(opts, "ship: git push origin %s:%s", out.Tip, branch)
	stdout, stderr, perr := gitRun(opts, hubDir, pushArgs...)
	combined := stdout + "\n" + stderr
	if perr != nil {
		verboseNotice(opts, "ship: MR push options rejected; fallback plain push")
		fallback := []string{"push", "origin", out.Tip + ":" + featureRef}
		if out.LeaseTip != "" {
			fallback = append(fallback, "--force-with-lease="+featureRef+":"+out.LeaseTip)
		}
		stdout2, stderr2, perr2 := gitRun(opts, hubDir, fallback...)
		if perr2 != nil {
			return fmt.Errorf("git push: %s", gitErrMsg(stderr2, perr2))
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
	return nil
}

// isRetryablePush reports whether a rejected publish may succeed after re-merging.
// A permission error or a bad refspec will not.
func isRetryablePush(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "non-fast-forward") ||
		// The ff-push refuses when the target no longer contains our tip: another
		// instance published between our merge and our push, which is the same race
		// as a rejected push and has the same fix.
		strings.Contains(msg, "not fast-forward") ||
		strings.Contains(msg, "fetch first") ||
		strings.Contains(msg, "cannot lock ref") ||
		strings.Contains(msg, "but expected") ||
		strings.Contains(msg, "updates were rejected")
}

// autoMergeFFPush fast-forwards origin/<target> to commitSHA without checking out
// target locally.
//
// It refuses when the target already contains the commit: the ff-merge is the
// last step of a ship whose feature branch is already published, so "nothing to
// do" must not be reported as a failure.
func autoMergeFFPush(opts Opts, hubDir, commitSHA, targetBranch string) error {
	if err := gitOK(opts, hubDir, "fetch", "origin", targetBranch); err != nil {
		return err
	}
	originTarget := "origin/" + targetBranch
	// A refusal here is a genuine "not fast-forward": this ff-push only ever runs on
	// the tip the merge produced, which by construction descends from the target.
	if err := gitOK(opts, hubDir, "merge-base", "--is-ancestor", originTarget, commitSHA); err != nil {
		return fmt.Errorf("not fast-forward: %s does not contain %s", originTarget, commitSHA)
	}
	refspec := commitSHA + ":refs/heads/" + targetBranch
	if err := gitOK(opts, hubDir, "push", "origin", refspec); err != nil {
		return fmt.Errorf("push origin/%s: %w", targetBranch, err)
	}
	_, _ = gitOut(opts, hubDir, "fetch", "origin", targetBranch)
	return nil
}

// defaultPublishAttempts bounds the merge/publish rounds when a peer keeps winning
// the race.
const defaultPublishAttempts = 3

// acquireShipLock takes the checkout-wide mutation lock for the git phase.
//
// The agent run is excluded: it is a minutes-long LLM call, and holding the lock
// across it would block every other Marcus writer. The lock covers commit →
// merge → resolve → push, which is where two publishers can actually collide.
func acquireShipLock(opts Opts, hubDir string) (func(), error) {
	if opts.LockFn == nil {
		return func() {}, nil
	}
	release, err := opts.LockFn(hubDir)
	if err != nil {
		return nil, fmt.Errorf("ship: repository busy (another Marcus operation holds the lock): %w", err)
	}
	if release == nil {
		return func() {}, nil
	}
	return release, nil
}

// syncBeforePublish merges origin/<target> into the hub and returns what it did.
// Nil is a legitimate result: a hub whose git cannot be probed keeps its previous
// behaviour rather than failing the ship for a detail.
func syncBeforePublish(opts Opts, hubDir, targetBranch string) (*ShipSyncResult, error) {
	if opts.SkipSync {
		return nil, nil
	}
	res, err := SyncHub(opts, hubDir, targetBranch)
	if err != nil {
		return res, err
	}
	if res != nil && len(res.Resolved) > 0 {
		verboseNotice(opts, "ship: resolver cleared %d conflicted path(s): %s",
			len(res.Resolved), strings.Join(res.Resolved, ", "))
	}
	return res, nil
}

// planSweep classifies the hub's remaining dirt so this commit can carry the
// knowledge that would otherwise sit uncommitted and block the next KB update.
//
// The paths result.json named are always committed, even if a binary: the agent
// wrote them deliberately and validating that intent is the caller's job, not a
// content sniff's.
// commitsAhead counts HEAD commits absent from ref. A git failure counts as zero:
// the number only feeds a log line.
func commitsAhead(opts Opts, hubDir, ref string) int {
	out, err := gitOut(opts, hubDir, "rev-list", "--count", ref+"..HEAD")
	if err != nil {
		return 0
	}
	n := 0
	_, _ = fmt.Sscanf(strings.TrimSpace(out), "%d", &n)
	return n
}

func planSweep(opts Opts, hubDir string, named []string) (swept []string, skipped []SweptPath) {
	if opts.SweepPlanFn == nil {
		return nil, nil
	}
	plan, err := opts.SweepPlanFn(hubDir)
	if err != nil || plan == nil {
		// A failed classification must not fail the ship: the named paths still go
		// in, and whatever is left over stays dirty exactly as before.
		return nil, nil
	}
	already := make(map[string]bool, len(named))
	for _, p := range named {
		already[filepath.ToSlash(strings.TrimSpace(p))] = true
	}
	for _, p := range plan.Include {
		rel := filepath.ToSlash(strings.TrimSpace(p.Path))
		if rel == "" || already[rel] {
			continue
		}
		swept = append(swept, rel)
	}
	for _, p := range plan.Exclude {
		rel := filepath.ToSlash(strings.TrimSpace(p.Path))
		if rel == "" || already[rel] {
			continue
		}
		skipped = append(skipped, SweptPath{Path: rel, Reason: p.Reason})
	}
	sort.Strings(swept)
	return swept, skipped
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
