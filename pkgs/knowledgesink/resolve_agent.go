package knowledgesink

import (
	"context"
	"fmt"
	"strings"
)

// runConflictAgent asks the configured runner to repair a conflicted merge.
//
// The agent edits files only. Staging, committing and the marker check stay here:
// a resolver that could also stage would be able to publish a half-resolved merge
// if it failed midway, and the caller would have no way to tell.
//
// This mirrors the repository-level on-conflict resolver contract in dot-pkgs
// (cwd = repo root, resolver never touches the index, bounded rounds); it differs
// in where the resolver comes from — an embedded skill plus the configured runner
// instead of a script committed in the repository, because the hub is reading
// material that a non-git user owns.
func runConflictAgent(ctx context.Context, opts Opts, dir string, conflicts []ConflictInfo) error {
	var list strings.Builder
	for _, c := range conflicts {
		list.WriteString("- " + c.Path)
		if c.Reason != "" {
			list.WriteString(" (" + c.Reason + ")")
		}
		list.WriteString("\n")
	}
	prompt := strings.TrimSpace(opts.ResolvePrompt) +
		"\n\nThe repository is " + dir + " and it is mid-merge. These paths are conflicted:\n\n" +
		list.String() +
		"\nResolve every one of them in place now. Do not run git add, git commit, " +
		"git merge --continue, git reset, or git checkout: the host stages and commits after you.\n"

	out, err := runAgent(ctx, opts, dir, prompt, "", "")
	if err != nil {
		return fmt.Errorf("conflict agent: %w", err)
	}
	if trimmed := strings.TrimSpace(out); trimmed != "" {
		// Anything the runner prints is diagnostic; the resolver's contract is the
		// working tree, not the output.
		verboseNotice(opts, "ship: conflict agent said: %s", firstLine(trimmed))
	}
	return nil
}

func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}
