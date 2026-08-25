package knowledgesink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ResultSchemaExample is the JSON object shape for headless propose-only RunJSON.
const ResultSchemaExample = `{
  "ok": true,
  "status": "proposed",
  "sink_index": 0,
  "proposal_path": "sink-0/proposal.md",
  "proposals": [
    {
      "path": "topics/example.md",
      "kind": "new",
      "rationale": "short why",
      "evidence": "pointer into session"
    }
  ],
  "error": ""
}`

// PromptInput fills AgentPrompt / ShowPromptText.
type PromptInput struct {
	MarcusSessionID  string
	Runner           string // grok-tty | codex-tty | …
	RunnerSessionID  string
	RunnerSessionDir string // absolute primary source
	Since            string // sunk cursor (last_sink_max_message_timestamp), optional
	SinkIndex        int
	ProposalPath     string // absolute path for proposal.md
	Prior            PriorSinkContext

	// CreateMR mode: agent applies hub writes and emits result.json for host git.
	CreateMR       bool
	ResultJSONPath string // absolute sink-N/result.json
	GitUser        string // email local-part for branch prefix
	BranchDate     string // YYYY-MM-DD
}

// PriorSinkContext is previous sink bookkeeping for dedup / incremental prompts.
type PriorSinkContext struct {
	LastSinkAt   string
	LastCursor   string
	PriorRunDirs []string // absolute sink-N dirs
	LastHubPaths []string
	LastPaths    []string // relative under session sink dir
	HasPrior     bool
}

func AgentPrompt(in PromptInput) string {
	if in.CreateMR {
		return agentPromptCreateMR(in)
	}
	return agentPromptProposeOnly(in)
}

func agentPromptProposeOnly(in PromptInput) string {
	var b strings.Builder
	b.WriteString(`You are proposing durable knowledge to sink into this knowledge-base-hub checkout.

CWD is the hub root. Read ./SINK.md first and follow it as the primary guideline.

Do NOT write, edit, or create hub files in this turn. Propose only.

Primary source (session transcript / artifacts) — treat this as ground truth:
`)
	writeSourceBlock(&b, in)
	writePriorBlock(&b, in)
	fmt.Fprintf(&b, `
Write your proposal markdown to this absolute path (outside the hub worktree):
  %s

Task:
1. Inspect the primary source (and only the new slice if a since/cursor is given).
2. Conclusion gate: if the session has no clear conclusion / decision / durable
   takeaway yet, write a short proposal noting that and stop — do not invent
   hub leaves from incomplete work.
3. Novelty gate: if nothing durable is new relative to the hub (and prior sinks),
   say so in the proposal and stop — empty proposals list is correct.
4. Otherwise, using SINK.md, propose what to add or adjust.
5. Do not re-propose knowledge that prior sinks already covered unless you are
   proposing a concrete incremental fix (path + what changes + why).
6. For each proposal: hub-relative target path, short rationale, evidence in the
   session, and whether it is new | adjust-existing.
7. Stop. Wait for human approval before any write.
`, in.ProposalPath)
	return strings.TrimSpace(b.String())
}

func agentPromptCreateMR(in PromptInput) string {
	var b strings.Builder
	b.WriteString(`You are applying durable knowledge into this knowledge-base-hub checkout for shipping.

CWD is the hub root. Read ./SINK.md first and follow it as the primary guideline.

Write, update, or delete hub files now (apply). Do NOT run git add, commit, or push — the host ships from result.json.
Classify every touched hub path under git_commit_files.add, .update, or .delete.

Primary source (session transcript / artifacts) — treat this as ground truth:
`)
	writeSourceBlock(&b, in)
	writePriorBlock(&b, in)
	date := strings.TrimSpace(in.BranchDate)
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	user := firstNonEmpty(strings.TrimSpace(in.GitUser), "user")
	fmt.Fprintf(&b, `
Write a short proposal markdown (audit) to this absolute path (outside the hub):
  %s

Task:
1. Inspect the primary source (and only the new slice if a since/cursor is given).
2. Conclusion gate: if the session has no clear conclusion / decision / durable
   takeaway yet → do NOT write hub files. Write result.json with
   has_new_knowledges=false and skip_reason="inconclusive", plus a short
   proposal.md explaining why. Stop.
3. Novelty gate: if nothing durable is new vs the hub (and prior sinks) → do
   NOT write hub files. Write result.json with has_new_knowledges=false and
   skip_reason="no_new", plus a short proposal.md. Stop.
4. Otherwise, using SINK.md, write the hub-relative files that should be sunk
   (new or adjust). Do not repeat prior sinks unless this is a concrete
   incremental fix.
5. Do not run git commands.

## Output
When finished, write JSON to this absolute path (outside the hub):
  %s

has_new_knowledges is required (true or false).
When true, ship fields are required and git_commit_files must list ≥1 path.
When false, set skip_reason to "inconclusive" or "no_new"; leave commit fields
empty and do not touch hub files.

Example (new knowledges):
%s

Example (skip):
%s

Branch format (only when has_new_knowledges=true): {user}/{YYYY-MM-DD}-{slug}
  user (from git): %s
  date: %s

git_commit_files is an object (not a string array):
  add:    new hub-relative files you created (must exist on disk)
  update: existing hub-relative files you modified (must exist on disk)
  delete: hub-relative files you removed (absent on disk, still tracked in git)
Omit empty buckets; at least one path overall is required iff has_new_knowledges=true.
`, in.ProposalPath, in.ResultJSONPath, ShipResultExample, ShipResultSkipExample, user, date)
	return strings.TrimSpace(b.String())
}

func writeSourceBlock(b *strings.Builder, in PromptInput) {
	fmt.Fprintf(b, "  runner: %s\n", firstNonEmpty(in.Runner, "unknown"))
	fmt.Fprintf(b, "  session_id: %s\n", in.RunnerSessionID)
	fmt.Fprintf(b, "  session_dir: %s\n", in.RunnerSessionDir)
	if strings.TrimSpace(in.Since) != "" {
		fmt.Fprintf(b, "  since: %s\n", in.Since)
	}
	fmt.Fprintf(b, "\nMarcus/agent-run id (bookkeeping): %s\n", firstNonEmpty(in.MarcusSessionID, "—"))
}

func writePriorBlock(b *strings.Builder, in PromptInput) {
	b.WriteString("\nPrevious sinks for this Marcus session (avoid repeating; prefer incremental updates):\n")
	if !in.Prior.HasPrior {
		b.WriteString("  (none yet — treat this as a first sink)\n")
		return
	}
	if in.Prior.LastSinkAt != "" {
		fmt.Fprintf(b, "  last_sink_at: %s\n", in.Prior.LastSinkAt)
	}
	if in.Prior.LastCursor != "" {
		fmt.Fprintf(b, "  last_cursor: %s\n", in.Prior.LastCursor)
	}
	if len(in.Prior.PriorRunDirs) > 0 {
		b.WriteString("  prior_run_dirs:\n")
		for _, d := range in.Prior.PriorRunDirs {
			fmt.Fprintf(b, "    - %s\n", d)
		}
	}
	if len(in.Prior.LastPaths) > 0 {
		b.WriteString("  last_paths:\n")
		for _, p := range in.Prior.LastPaths {
			fmt.Fprintf(b, "    - %s\n", p)
		}
	}
	if len(in.Prior.LastHubPaths) > 0 {
		b.WriteString("  last_hub_paths:\n")
		for _, p := range in.Prior.LastHubPaths {
			fmt.Fprintf(b, "    - %s\n", p)
		}
	}
}

// ShowPromptText is the reviewable prompt (no disk writes).
func ShowPromptText(in PromptInput) string {
	if strings.TrimSpace(in.ProposalPath) == "" {
		in.ProposalPath = fmt.Sprintf("<would-be sink-%d/proposal.md>", in.SinkIndex)
	}
	if in.CreateMR && strings.TrimSpace(in.ResultJSONPath) == "" {
		in.ResultJSONPath = fmt.Sprintf("<would-be sink-%d/result.json>", in.SinkIndex)
	}
	return AgentPrompt(in)
}

func buildPriorContext(stateSessionDir string, manifest *Manifest) PriorSinkContext {
	out := PriorSinkContext{}
	if manifest == nil {
		return out
	}
	has := strings.TrimSpace(manifest.LastSinkAt) != "" ||
		SunkCursor(manifest) != "" ||
		CheckedCursor(manifest) != "" ||
		manifest.LastSinkIndex >= 0 ||
		len(manifest.LastPaths) > 0 ||
		len(manifest.LastHubPaths) > 0 ||
		manifest.NextSinkIndex > 0
	if !has {
		return out
	}
	out.HasPrior = true
	out.LastSinkAt = manifest.LastSinkAt
	out.LastCursor = SunkCursor(manifest)
	out.LastPaths = append([]string(nil), manifest.LastPaths...)
	out.LastHubPaths = append([]string(nil), manifest.LastHubPaths...)
	max := manifest.NextSinkIndex
	if manifest.LastSinkIndex+1 > max {
		max = manifest.LastSinkIndex + 1
	}
	for i := 0; i < max; i++ {
		d := RunDir(stateSessionDir, i)
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			out.PriorRunDirs = append(out.PriorRunDirs, d)
		}
	}
	return out
}

func proposalRelPath(index int) string {
	return filepath.ToSlash(filepath.Join(fmt.Sprintf("sink-%d", index), "proposal.md"))
}

func formatProposalMarkdown(in PromptInput, proposals []proposalItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Sink proposal %d\n\n", in.SinkIndex)
	fmt.Fprintf(&b, "- marcus_session_id: %s\n", in.MarcusSessionID)
	fmt.Fprintf(&b, "- runner: %s\n", in.Runner)
	fmt.Fprintf(&b, "- runner_session_id: %s\n", in.RunnerSessionID)
	fmt.Fprintf(&b, "- runner_session_dir: %s\n", in.RunnerSessionDir)
	if in.Since != "" {
		fmt.Fprintf(&b, "- since: %s\n", in.Since)
	}
	if in.CreateMR {
		fmt.Fprintf(&b, "- status: applied\n\n")
	} else {
		fmt.Fprintf(&b, "- status: proposed\n\n")
	}
	b.WriteString("## Proposals\n\n")
	if len(proposals) == 0 {
		b.WriteString("(none)\n")
		return b.String()
	}
	for i, p := range proposals {
		fmt.Fprintf(&b, "### %d. %s (%s)\n\n", i+1, firstNonEmpty(p.Path, "—"), firstNonEmpty(p.Kind, "new"))
		if p.Rationale != "" {
			fmt.Fprintf(&b, "%s\n\n", p.Rationale)
		}
		if p.Evidence != "" {
			fmt.Fprintf(&b, "evidence: %s\n\n", p.Evidence)
		}
	}
	return b.String()
}

type proposalItem struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"` // new | adjust-existing
	Rationale string `json:"rationale"`
	Evidence  string `json:"evidence"`
}
