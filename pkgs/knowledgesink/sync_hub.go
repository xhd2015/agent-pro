package knowledgesink

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/merge"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/sweep"
)

// ShipSyncResult is what bringing origin/<target> into the hub did.
type ShipSyncResult struct {
	Tip      string
	Synced   bool
	Incoming int
	Resolved []string
	Foreign  int
}

// SyncHub merges origin/<target> into the hub checkout, resolving conflicts with
// the configured prompt.
//
// Merge, never rebase: the hub is a shared integration branch that every sink
// publishes into, so rewriting its commits would break everyone else's clone.
//
// Conflicts are rare and are always the same shape — two people edited the same
// knowledge page — so they are handed to an agent that has the repository's
// resolution rules, rather than failed back to a user who cannot read a conflict
// marker.
func SyncHub(opts Opts, hubDir, targetBranch string) (*ShipSyncResult, error) {
	hooks, err := newHooksSuppressor()
	if err != nil {
		return nil, err
	}
	defer hooks.release()

	res, err := merge.Sync(context.Background(), merge.Options{
		Dir:            hubDir,
		Remote:         "origin",
		Branch:         targetBranch,
		Message:        "chore(kb): merge origin/" + targetBranch,
		Push:           false, // ShipToMR publishes the tip itself, to both refs.
		PublishForeign: true,
		Resolve:        conflictResolver(opts),
		Config:         []string{"core.hooksPath=" + hooks.hooksDir},
		// One merge has a single stop; retrying the resolver only helps when it was
		// interrupted, so two attempts is already generous.
		MaxAttempts: 1,
	}, opts.Stderr)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return &ShipSyncResult{
		Tip:      res.Tip,
		Synced:   res.Synced,
		Incoming: res.Incoming,
		Resolved: res.Resolved,
		Foreign:  res.Foreign,
	}, nil
}

// SweepPlan is the classification the spl CLI injects. It is the same predicate
// that decides whether the KB may be updated, so "committed" and "clean enough to
// update" can never disagree.
type SweepPlan = sweep.Plan

// NewConflictResolver builds the resolver handed to merge.Sync, for callers that
// drive the merge themselves (the daily `marcus kb sweep`, which is a CLI process
// rather than a sink run).
//
// It returns nil when no briefing or no agent is configured: a conflicted merge is
// then aborted instead of guessed at.
func NewConflictResolver(opts Opts) merge.ResolveFunc { return conflictResolver(opts) }

// conflictResolver builds the resolver handed to merge.Sync.
//
// Without a prompt there is no resolver: the merge is aborted and the ship fails
// with the conflicted paths, which is the safe answer when nobody configured how
// to resolve.
//
// Only the prompt is required. An injected AgentFn is a test seam, not a
// precondition: production runs reach the runner through runAgent like every other
// agent call in this package.
func conflictResolver(opts Opts) merge.ResolveFunc {
	if strings.TrimSpace(opts.ResolvePrompt) == "" {
		return nil
	}
	return func(ctx context.Context, dir string, conflicts []ConflictInfo) error {
		return runConflictAgent(ctx, opts, dir, conflicts)
	}
}

// ConflictInfo mirrors merge.Conflict so the resolver signature can be aliased
// without leaking the dot-pkgs type into every caller.
type ConflictInfo = merge.Conflict

// newHooksSuppressor creates the empty hooks directory both merge and the ship
// use, so a developer's global hooks cannot fail an unattended commit.
type hooksSuppressor struct {
	hooksDir string
}

func newHooksSuppressor() (*hooksSuppressor, error) {
	dir, err := os.MkdirTemp("", "knowledgesink-merge-hooks-*")
	if err != nil {
		return nil, fmt.Errorf("mkdir hooks stub: %w", err)
	}
	return &hooksSuppressor{hooksDir: dir}, nil
}

func (h *hooksSuppressor) release() {
	if h != nil && h.hooksDir != "" {
		_ = os.RemoveAll(h.hooksDir)
	}
}
