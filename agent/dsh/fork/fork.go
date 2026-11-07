// Package fork launches a DSH session as a child of another session in a
// different checkout. The child inherits the source transcript up to a chosen
// cut, so an Agent can keep its context while working in a new worktree without
// disturbing the conversation that is still running.
//
// The package owns the fork policy — checkout separation, resume of a partially
// delivered child, verification that the Host honoured the checkout, draft
// seeding and the optional browser launch. It never writes the session store:
// every fact about a session comes from the caller-supplied reader.
package fork

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/dsh/web"
)

// DefaultTimeout bounds one attempt: a long transcript can take a while to seed,
// and the Host composes the child Agent before it answers.
const DefaultTimeout = 2 * time.Minute

// SessionInfo is the durable or live fact a fork needs about one session. Cwd is
// the session's working directory as recorded by the Host; an empty Cwd means
// the reader cannot tell and skips the checkout checks.
type SessionInfo struct {
	ID  string
	Cwd string
}

// Refusals a caller should surface verbatim; errors.Is matches all of them.
var (
	// ErrSourceMissing means the fork source is not a known session.
	ErrSourceMissing = errors.New("source session not found")
	// ErrSameCheckout means the child would share the source checkout and could
	// interfere with the Agent that is still running there.
	ErrSameCheckout = errors.New("fork needs a checkout different from the source session")
	// ErrChildElsewhere means a resumed child belongs to a different checkout.
	ErrChildElsewhere = errors.New("session does not belong to the selected checkout")
	// ErrCwdNotHonored means the Host forked into the source checkout, which only
	// a Host without fork-checkout support does.
	ErrCwdNotHonored = errors.New("fork stayed in the source checkout; update the harness that serves dsh web")
)

// Options is one fork attempt.
type Options struct {
	// Client talks to the Host. Required; a nil HTTPClient gets one bounded by
	// Timeout, because the default is too short for seeding a long transcript.
	Client web.Client
	// ReadSession reports a session's current checkout. Required; pass the
	// caller's live-aware reader so a running Agent is seen.
	ReadSession func(id string) (SessionInfo, error)
	// SourceID is the session whose transcript the child inherits. Required.
	SourceID string
	// Cwd is the child checkout, which must differ from the source's. Required.
	Cwd string
	// AtSeq is the inclusive transcript cut; 0 keeps the latest completed turn.
	AtSeq int64
	// Resume is a child ID from an earlier partial attempt; when it still
	// belongs to Cwd the fork is skipped and only the remaining steps run.
	Resume string
	// Draft is stored in the child composer for review and never submitted.
	Draft string
	// Browser selects the window opened when Open is set; web.BrowserNone skips it.
	Browser string
	// Open launches the child's browser window after the draft is seeded.
	Open bool
	// Timeout bounds the attempt; 0 uses DefaultTimeout.
	Timeout time.Duration
}

// Result reports what the attempt did. SessionID is set as soon as a child
// exists, even when a later step fails, so a caller can checkpoint it and retry
// with Resume.
type Result struct {
	SessionID string
	Forked    bool
	Drafted   bool
	Opened    bool
}

// Run forks SourceID into Cwd and seeds the draft. A refused fork returns no
// session ID; a failure after the fork keeps it so the caller can retry.
func Run(ctx context.Context, opts Options) (Result, error) {
	if opts.ReadSession == nil {
		return Result{}, fmt.Errorf("DSH fork requires a session reader")
	}
	sourceID := strings.TrimSpace(opts.SourceID)
	if sourceID == "" {
		return Result{}, fmt.Errorf("DSH fork requires a source session ID")
	}
	cwd, err := checkCheckout(opts.Cwd)
	if err != nil {
		return Result{}, err
	}
	if opts.AtSeq < 0 {
		return Result{}, fmt.Errorf("DSH fork cut must not be negative")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	client := opts.Client
	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{Timeout: timeout}
	}

	source, err := opts.ReadSession(sourceID)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrSourceMissing, err)
	}
	if source.Cwd != "" && samePath(source.Cwd, cwd) {
		return Result{}, ErrSameCheckout
	}
	result := Result{}
	if resume := strings.TrimSpace(opts.Resume); resume != "" {
		// A retry may reopen the child returned by a partial failure, but must
		// never adopt a session that belongs to another checkout.
		child, err := opts.ReadSession(resume)
		if err != nil || !samePath(child.Cwd, cwd) {
			return Result{}, fmt.Errorf("%w: %s", ErrChildElsewhere, resume)
		}
		result.SessionID = resume
	} else {
		child, err := client.Fork(ctx, web.ForkOptions{SessionID: sourceID, AtSeq: opts.AtSeq, Cwd: cwd})
		if err != nil {
			return Result{}, err
		}
		result.SessionID, result.Forked = child, true
		// A Host without fork-checkout support drops the field and silently forks
		// into the source directory; refuse that child instead of sharing it.
		if childInfo, err := opts.ReadSession(child); err == nil && childInfo.Cwd != "" && !samePath(childInfo.Cwd, cwd) {
			return result, fmt.Errorf("%w: %s", ErrCwdNotHonored, child)
		}
	}
	if strings.TrimSpace(opts.Draft) != "" {
		if err := client.SetDraft(ctx, result.SessionID, opts.Draft); err != nil {
			return result, fmt.Errorf("DSH fork draft: %w", err)
		}
		result.Drafted = true
	}
	if opts.Open {
		if err := client.Open(result.SessionID, opts.Browser); err != nil {
			return result, fmt.Errorf("DSH fork open: %w", err)
		}
		result.Opened = true
	}
	return result, nil
}

// checkCheckout requires a fully qualified existing directory: the Host matches
// the checkout against its workspace registry and keeps the source for anything
// else, so a typo must fail before the fork.
func checkCheckout(raw string) (string, error) {
	cwd := strings.TrimSpace(raw)
	if cwd == "" {
		return "", fmt.Errorf("DSH fork requires a checkout directory")
	}
	if !filepath.IsAbs(cwd) {
		return "", fmt.Errorf("fork cwd must be a fully qualified directory")
	}
	st, err := os.Stat(cwd)
	if err != nil {
		return "", fmt.Errorf("fork cwd: %w", err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("fork cwd is not a directory")
	}
	return cwd, nil
}

// samePath treats symlinked and /private-prefixed spellings of one directory as
// equal, which is what the Host's own workspace matching does.
func samePath(a, b string) bool {
	na, nb := canonicalPath(a), canonicalPath(b)
	if na == "" || nb == "" {
		return na == nb
	}
	return na == nb
}

func canonicalPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = filepath.Clean(abs)
		}
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return filepath.Clean(resolved)
	}
	return path
}
