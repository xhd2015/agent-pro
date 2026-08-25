package knowledgesink

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Skip reasons when has_new_knowledges is false.
const (
	SkipReasonInconclusive = "inconclusive" // no clear conclusion yet — advance checked only
	SkipReasonNoNew        = "no_new"       // nothing new vs hub — advance checked only
)

// ShipResultExample is the example object for the prompt Output section (new knowledges).
const ShipResultExample = `{
  "has_new_knowledges": true,
  "git_commit_msg": "docs(kb): sink session learnings",
  "git_branch_name": "devuser/2026-03-24-short-slug",
  "git_commit_files": {
    "add": ["topics/example.md"],
    "update": ["INDEX.md"],
    "delete": []
  }
}`

// ShipResultSkipExample is the skip contract (no hub writes).
const ShipResultSkipExample = `{
  "has_new_knowledges": false,
  "skip_reason": "no_new",
  "git_commit_msg": "",
  "git_branch_name": "",
  "git_commit_files": {}
}`

// ShipCommitFiles is the agent-written hub path set for --create-mr shipping.
// add/update paths must exist on disk; delete paths must be absent and still
// tracked in git. Empty buckets may be omitted; at least one path overall
// when has_new_knowledges is true.
type ShipCommitFiles struct {
	Add    []string `json:"add,omitempty"`
	Update []string `json:"update,omitempty"`
	Delete []string `json:"delete,omitempty"`
}

// UnmarshalJSON rejects legacy flat string arrays with a clear contract error.
func (f *ShipCommitFiles) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = ShipCommitFiles{}
		return nil
	}
	if data[0] == '[' {
		return fmt.Errorf("git_commit_files must be object {\"add\":[],\"update\":[],\"delete\":[]}, not a string array")
	}
	type alias ShipCommitFiles
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*f = ShipCommitFiles(a)
	return nil
}

// AllPaths returns add ∪ update ∪ delete in that order (no dedupe across buckets).
func (f ShipCommitFiles) AllPaths() []string {
	n := len(f.Add) + len(f.Update) + len(f.Delete)
	if n == 0 {
		return nil
	}
	out := make([]string, 0, n)
	out = append(out, f.Add...)
	out = append(out, f.Update...)
	out = append(out, f.Delete...)
	return out
}

// ShipResult is the agent-written contract for --create-mr shipping.
type ShipResult struct {
	// HasNewKnowledges is required. Pointer so missing JSON is distinct from false.
	HasNewKnowledges *bool           `json:"has_new_knowledges"`
	SkipReason       string          `json:"skip_reason,omitempty"` // inconclusive|no_new when false
	GitCommitMsg     string          `json:"git_commit_msg"`
	GitBranchName    string          `json:"git_branch_name"`
	GitCommitFiles   ShipCommitFiles `json:"git_commit_files"`
}

// HasNew reports whether the agent marked new knowledges (false if nil/missing).
func (sr *ShipResult) HasNew() bool {
	return sr != nil && sr.HasNewKnowledges != nil && *sr.HasNewKnowledges
}

// BoolPtr returns a *bool for ShipResult literals in tests and helpers.
func BoolPtr(v bool) *bool { return &v }

var branchNameRe = regexp.MustCompile(`^[^/\s]+/\d{4}-\d{2}-\d{2}-[a-z0-9][a-z0-9-]*$`)

// UserFromEmail strips the domain from an email (devuser@example.com → devuser).
func UserFromEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
}

func resultJSONRelPath(index int) string {
	return filepath.ToSlash(filepath.Join(fmt.Sprintf("sink-%d", index), "result.json"))
}

func resultJSONAbsPath(stateSessionDir string, index int) string {
	return filepath.Join(RunDir(stateSessionDir, index), "result.json")
}

// ReadValidateShipResult loads and validates sink-N/result.json against hubDir.
func ReadValidateShipResult(path, hubDir string) (*ShipResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("result.json missing or malformed at %s", path)
		}
		return nil, fmt.Errorf("result.json missing or malformed at %s: %w", path, err)
	}
	var sr ShipResult
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, fmt.Errorf("result.json missing or malformed at %s: %w", path, err)
	}
	if err := validateShipResult(&sr, hubDir); err != nil {
		return nil, err
	}
	return &sr, nil
}

type pathOp string

const (
	pathOpAdd    pathOp = "add"
	pathOpUpdate pathOp = "update"
	pathOpDelete pathOp = "delete"
)

func validateShipResult(sr *ShipResult, hubDir string) error {
	if sr == nil {
		return fmt.Errorf("result.json missing or malformed: empty")
	}
	if sr.HasNewKnowledges == nil {
		return fmt.Errorf("result.json: has_new_knowledges is required (true or false)")
	}

	if !*sr.HasNewKnowledges {
		reason := strings.ToLower(strings.TrimSpace(sr.SkipReason))
		switch reason {
		case SkipReasonInconclusive, SkipReasonNoNew:
			sr.SkipReason = reason
		default:
			return fmt.Errorf("result.json: skip_reason must be %q or %q when has_new_knowledges=false (got %q)",
				SkipReasonInconclusive, SkipReasonNoNew, sr.SkipReason)
		}
		if len(sr.GitCommitFiles.AllPaths()) > 0 {
			return fmt.Errorf("result.json: has_new_knowledges=false must not list git_commit_files paths")
		}
		sr.GitCommitMsg = strings.TrimSpace(sr.GitCommitMsg)
		sr.GitBranchName = strings.TrimSpace(sr.GitBranchName)
		sr.GitCommitFiles = ShipCommitFiles{}
		return nil
	}

	msg := strings.TrimSpace(sr.GitCommitMsg)
	branch := strings.TrimSpace(sr.GitBranchName)
	if msg == "" {
		return fmt.Errorf("result.json: git_commit_msg is required when has_new_knowledges=true")
	}
	if branch == "" {
		return fmt.Errorf("result.json: git_branch_name is required when has_new_knowledges=true")
	}
	if !branchNameRe.MatchString(strings.ToLower(branch)) {
		if !looseBranchOK(branch) {
			return fmt.Errorf("result.json: git_branch_name must look like user/YYYY-MM-DD-slug (got %q)", branch)
		}
	}

	hubAbs, err := filepath.Abs(hubDir)
	if err != nil {
		return fmt.Errorf("hub path: %w", err)
	}

	seen := map[string]pathOp{}
	cleanAdd, err := validatePathBucket(sr.GitCommitFiles.Add, pathOpAdd, hubAbs, seen)
	if err != nil {
		return err
	}
	cleanUpdate, err := validatePathBucket(sr.GitCommitFiles.Update, pathOpUpdate, hubAbs, seen)
	if err != nil {
		return err
	}
	cleanDelete, err := validatePathBucket(sr.GitCommitFiles.Delete, pathOpDelete, hubAbs, seen)
	if err != nil {
		return err
	}
	if len(cleanAdd)+len(cleanUpdate)+len(cleanDelete) == 0 {
		return fmt.Errorf("result.json: has_new_knowledges=true requires at least one path in git_commit_files")
	}

	sr.SkipReason = ""
	sr.GitCommitMsg = msg
	sr.GitBranchName = branch
	sr.GitCommitFiles = ShipCommitFiles{Add: cleanAdd, Update: cleanUpdate, Delete: cleanDelete}
	return nil
}

func validatePathBucket(raw []string, op pathOp, hubAbs string, seen map[string]pathOp) ([]string, error) {
	cleaned := make([]string, 0, len(raw))
	bucketSeen := map[string]struct{}{}
	for _, rawPath := range raw {
		p := filepath.ToSlash(strings.TrimSpace(rawPath))
		if p == "" {
			return nil, fmt.Errorf("result.json: empty path in git_commit_files.%s", op)
		}
		if filepath.IsAbs(p) || strings.HasPrefix(p, "/") {
			return nil, fmt.Errorf("result.json: path must be hub-relative: %s", p)
		}
		if strings.Contains(p, "..") {
			return nil, fmt.Errorf("result.json: path escapes hub: %s", p)
		}
		full := filepath.Join(hubAbs, filepath.FromSlash(p))
		fullAbs, aerr := filepath.Abs(full)
		if aerr != nil {
			return nil, fmt.Errorf("result.json: resolve %s: %w", p, aerr)
		}
		rel, rerr := filepath.Rel(hubAbs, fullAbs)
		if rerr != nil || strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("result.json: path escapes hub: %s", p)
		}
		norm := filepath.ToSlash(rel)
		if prev, ok := seen[norm]; ok {
			return nil, fmt.Errorf("result.json: path %s listed in both %s and %s", norm, prev, op)
		}
		if _, ok := bucketSeen[norm]; ok {
			continue
		}

		fi, serr := os.Stat(fullAbs)
		switch op {
		case pathOpAdd, pathOpUpdate:
			if serr != nil {
				if os.IsNotExist(serr) {
					return nil, fmt.Errorf("result.json: file missing after agent (%s): %s", op, norm)
				}
				return nil, fmt.Errorf("result.json: resolve %s: %w", norm, serr)
			}
			if fi.IsDir() {
				return nil, fmt.Errorf("result.json: file missing after agent (%s): %s", op, norm)
			}
		case pathOpDelete:
			if serr == nil {
				return nil, fmt.Errorf("result.json: delete path still exists on disk: %s", norm)
			}
			if !os.IsNotExist(serr) {
				return nil, fmt.Errorf("result.json: resolve %s: %w", norm, serr)
			}
			if !hubPathTracked(hubAbs, norm) {
				return nil, fmt.Errorf("result.json: delete path not tracked in git: %s", norm)
			}
		default:
			return nil, fmt.Errorf("result.json: unknown path op %q", op)
		}

		bucketSeen[norm] = struct{}{}
		seen[norm] = op
		cleaned = append(cleaned, norm)
	}
	return cleaned, nil
}

// hubPathTracked reports whether rel is in the hub git index/HEAD (worktree may be deleted).
func hubPathTracked(hubDir, rel string) bool {
	out, _, err := defaultGitRunner(hubDir, "ls-files", "--", rel)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if filepath.ToSlash(strings.TrimSpace(line)) == filepath.ToSlash(rel) {
			return true
		}
	}
	return false
}

func looseBranchOK(branch string) bool {
	parts := strings.Split(branch, "/")
	if len(parts) < 2 {
		return false
	}
	user := parts[0]
	rest := strings.Join(parts[1:], "/")
	if strings.TrimSpace(user) == "" || strings.ContainsAny(user, " \t") {
		return false
	}
	return branchNameRe.MatchString(strings.ToLower(user + "/" + rest))
}

// Marcus sink trigger sources (CLI --source / daemon POST source).
const (
	SourceAuto  = "auto"
	SourceUI    = "ui"
	SourceSlash = "slash"
)

// SingleLineMRTitle flattens a commit message for merge_request.title push option.
func SingleLineMRTitle(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "knowledge sink"
	}
	if i := strings.IndexAny(msg, "\r\n"); i >= 0 {
		msg = strings.TrimSpace(msg[:i])
	}
	msg = strings.ReplaceAll(msg, ",", " ")
	if msg == "" {
		return "knowledge sink"
	}
	return msg
}

// MRTitlePrefix returns the Marcus trigger tag(s) for an MR title, or "" when
// source is empty/unknown (bare CLI).
func MRTitlePrefix(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case SourceAuto:
		return "[Auto Sink]"
	case SourceUI:
		return "[Auto Sink] [From UI]"
	case SourceSlash:
		return "[Auto Sink] [From /sink]"
	default:
		return ""
	}
}

// FormatMRTitle builds merge_request.title: optional source prefix + agent line.
func FormatMRTitle(source, commitMsg string) string {
	base := SingleLineMRTitle(commitMsg)
	prefix := MRTitlePrefix(source)
	if prefix == "" {
		return base
	}
	return strings.TrimSpace(prefix + " " + base)
}
