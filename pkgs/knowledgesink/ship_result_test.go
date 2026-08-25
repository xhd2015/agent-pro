package knowledgesink

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserFromEmail(t *testing.T) {
	if got := UserFromEmail("devuser@example.com"); got != "devuser" {
		t.Fatalf("got %q", got)
	}
	if got := UserFromEmail("  a@b.c  "); got != "a" {
		t.Fatalf("got %q", got)
	}
	if got := UserFromEmail("nodomain"); got != "nodomain" {
		t.Fatalf("got %q", got)
	}
}

func TestSingleLineMRTitle(t *testing.T) {
	if got := SingleLineMRTitle("one\ntwo"); got != "one" {
		t.Fatalf("got %q", got)
	}
}

func TestMRTitlePrefixAndFormat(t *testing.T) {
	cases := []struct {
		source string
		msg    string
		want   string
	}{
		{"auto", "docs(kb): learn\nbody", "[Auto Sink] docs(kb): learn"},
		{"ui", "docs(kb): learn", "[Auto Sink] [From UI] docs(kb): learn"},
		{"slash", "docs(kb): learn", "[Auto Sink] [From /sink] docs(kb): learn"},
		{"", "docs(kb): learn", "docs(kb): learn"},
		{"other", "docs(kb): learn", "docs(kb): learn"},
		{"ui", "", "[Auto Sink] [From UI] knowledge sink"},
	}
	for _, tc := range cases {
		if got := FormatMRTitle(tc.source, tc.msg); got != tc.want {
			t.Errorf("source=%q msg=%q: got %q want %q", tc.source, tc.msg, got, tc.want)
		}
	}
	if MRTitlePrefix("UI") != "[Auto Sink] [From UI]" {
		t.Fatalf("case-insensitive UI prefix failed")
	}
}

func TestShipCommitFilesAllPaths(t *testing.T) {
	f := ShipCommitFiles{
		Add:    []string{"a.md"},
		Update: []string{"b.md"},
		Delete: []string{"c.md"},
	}
	got := f.AllPaths()
	want := []string{"a.md", "b.md", "c.md"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestValidateShipResult(t *testing.T) {
	hub := t.TempDir()
	if err := os.MkdirAll(filepath.Join(hub, "topics"), 0o755); err != nil {
		t.Fatal(err)
	}
	leaf := filepath.Join(hub, "topics", "a.md")
	if err := os.WriteFile(leaf, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ok := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): x",
		GitBranchName:    "devuser/2026-03-24-sink-x",
		GitCommitFiles: ShipCommitFiles{
			Update: []string{"topics/a.md"},
		},
	}
	if err := validateShipResult(ok, hub); err != nil {
		t.Fatal(err)
	}

	bad := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "m",
		GitBranchName:    "devuser/2026-03-24-sink-x",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"../etc/passwd"},
		},
	}
	if err := validateShipResult(bad, hub); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("err = %v", err)
	}

	missing := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "m",
		GitBranchName:    "devuser/2026-03-24-sink-x",
		GitCommitFiles: ShipCommitFiles{
			Add: []string{"topics/missing.md"},
		},
	}
	if err := validateShipResult(missing, hub); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateShipResultAllowsDeletedTrackedFile(t *testing.T) {
	hub, _ := setupHubRemote(t)
	gone := filepath.Join(hub, "README.md")
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): drop README",
		GitBranchName:    "tester/2026-03-24-drop-readme",
		GitCommitFiles: ShipCommitFiles{
			Update: []string{"SINK.md"},
			Delete: []string{"README.md"},
		},
	}
	if err := validateShipResult(sr, hub); err != nil {
		t.Fatal(err)
	}
	if len(sr.GitCommitFiles.Delete) != 1 || sr.GitCommitFiles.Delete[0] != "README.md" {
		t.Fatalf("files = %+v", sr.GitCommitFiles)
	}
}

func TestValidateShipResultRejectsDeleteStillPresent(t *testing.T) {
	hub, _ := setupHubRemote(t)
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): bad delete",
		GitBranchName:    "tester/2026-03-24-bad-delete",
		GitCommitFiles: ShipCommitFiles{
			Delete: []string{"README.md"},
		},
	}
	if err := validateShipResult(sr, hub); err == nil || !strings.Contains(err.Error(), "still exists") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateShipResultRejectsMissingUntrackedDelete(t *testing.T) {
	hub, _ := setupHubRemote(t)
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): invent",
		GitBranchName:    "tester/2026-03-24-invent",
		GitCommitFiles: ShipCommitFiles{
			Delete: []string{"never-existed.md"},
		},
	}
	if err := validateShipResult(sr, hub); err == nil || !strings.Contains(err.Error(), "not tracked") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateShipResultRejectsCrossBucketDuplicate(t *testing.T) {
	hub := t.TempDir()
	if err := os.WriteFile(filepath.Join(hub, "a.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): dup",
		GitBranchName:    "tester/2026-03-24-dup",
		GitCommitFiles: ShipCommitFiles{
			Add:    []string{"a.md"},
			Update: []string{"a.md"},
		},
	}
	if err := validateShipResult(sr, hub); err == nil || !strings.Contains(err.Error(), "both") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateShipResultRejectsEmptyBuckets(t *testing.T) {
	hub := t.TempDir()
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): empty",
		GitBranchName:    "tester/2026-03-24-empty",
		GitCommitFiles:   ShipCommitFiles{},
	}
	if err := validateShipResult(sr, hub); err == nil || !strings.Contains(err.Error(), "has_new_knowledges=true requires at least one path") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateShipResultRequiresHasNewKnowledges(t *testing.T) {
	hub := t.TempDir()
	sr := &ShipResult{
		GitCommitMsg:  "docs(kb): x",
		GitBranchName: "tester/2026-03-24-x",
		GitCommitFiles: ShipCommitFiles{
			Update: []string{"a.md"},
		},
	}
	if err := validateShipResult(sr, hub); err == nil || !strings.Contains(err.Error(), "has_new_knowledges is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateShipResultSkipNoNew(t *testing.T) {
	hub := t.TempDir()
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(false),
		SkipReason:       SkipReasonNoNew,
	}
	if err := validateShipResult(sr, hub); err != nil {
		t.Fatal(err)
	}
	if sr.SkipReason != SkipReasonNoNew {
		t.Fatalf("skip=%q", sr.SkipReason)
	}
}

func TestValidateShipResultSkipInconclusive(t *testing.T) {
	hub := t.TempDir()
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(false),
		SkipReason:       "Inconclusive",
	}
	if err := validateShipResult(sr, hub); err != nil {
		t.Fatal(err)
	}
	if sr.SkipReason != SkipReasonInconclusive {
		t.Fatalf("skip=%q", sr.SkipReason)
	}
}

func TestValidateShipResultSkipRequiresReason(t *testing.T) {
	hub := t.TempDir()
	sr := &ShipResult{HasNewKnowledges: BoolPtr(false)}
	if err := validateShipResult(sr, hub); err == nil || !strings.Contains(err.Error(), "skip_reason") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateShipResultSkipRejectsPaths(t *testing.T) {
	hub := t.TempDir()
	if err := os.WriteFile(filepath.Join(hub, "a.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sr := &ShipResult{
		HasNewKnowledges: BoolPtr(false),
		SkipReason:       SkipReasonNoNew,
		GitCommitFiles:   ShipCommitFiles{Update: []string{"a.md"}},
	}
	if err := validateShipResult(sr, hub); err == nil || !strings.Contains(err.Error(), "must not list") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadValidateShipResult(t *testing.T) {
	hub := t.TempDir()
	if err := os.WriteFile(filepath.Join(hub, "INDEX.md"), []byte("#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "result.json")
	body, _ := json.Marshal(ShipResult{
		HasNewKnowledges: BoolPtr(true),
		GitCommitMsg:     "docs(kb): i",
		GitBranchName:    "u/2026-01-02-slug",
		GitCommitFiles: ShipCommitFiles{
			Update: []string{"INDEX.md"},
		},
	})
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	sr, err := ReadValidateShipResult(path, hub)
	if err != nil || sr == nil || sr.GitBranchName != "u/2026-01-02-slug" {
		t.Fatalf("%v %+v", err, sr)
	}
	if _, err := ReadValidateShipResult(filepath.Join(t.TempDir(), "nope.json"), hub); err == nil {
		t.Fatal("expected missing error")
	}
}

func TestReadValidateShipResultRejectsLegacyArray(t *testing.T) {
	hub := t.TempDir()
	if err := os.WriteFile(filepath.Join(hub, "INDEX.md"), []byte("#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "result.json")
	body := []byte(`{
  "git_commit_msg": "docs(kb): legacy",
  "git_branch_name": "u/2026-01-02-slug",
  "git_commit_files": ["INDEX.md"]
}`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadValidateShipResult(path, hub)
	if err == nil || !strings.Contains(err.Error(), "must be object") {
		t.Fatalf("err = %v", err)
	}
}

func TestAgentPromptCreateMRHasOutput(t *testing.T) {
	p := AgentPrompt(PromptInput{
		MarcusSessionID:  "m1",
		Runner:           "grok-tty",
		RunnerSessionID:  "g1",
		RunnerSessionDir: "/tmp/sess",
		SinkIndex:        2,
		ProposalPath:     "/tmp/state/sink-2/proposal.md",
		CreateMR:         true,
		ResultJSONPath:   "/tmp/state/sink-2/result.json",
		GitUser:          "devuser",
		BranchDate:       "2026-03-24",
	})
	for _, want := range []string{
		"./SINK.md",
		"## Output",
		"/tmp/state/sink-2/result.json",
		"has_new_knowledges",
		"skip_reason",
		"Golden rules",
		"Do not sink an explanation of what current code already does.",
		"reusable investigation value",
		"pitfall that can mislead future diagnosis",
		"correct path/evidence",
		"feature-specific leaf",
		"inconclusive",
		"no_new",
		"git_commit_msg",
		"git_branch_name",
		"git_commit_files",
		`"add":`,
		`"update":`,
		`"delete":`,
		"devuser",
		"2026-03-24",
		"Do NOT run git",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("prompt missing %q:\n%s", want, p)
		}
	}
	if strings.Contains(p, "Do NOT write, edit, or create hub files") {
		t.Fatal("create-mr prompt must allow hub writes")
	}
	if strings.Contains(p, `"git_commit_files": [`) {
		t.Fatal("prompt example must not use legacy array form")
	}
}

func TestOriginToWebBase(t *testing.T) {
	if got := originToWebBase("git@gitlab.example.com:group/proj.git"); got != "https://gitlab.example.com/group/proj" {
		t.Fatalf("got %q", got)
	}
	if got := originToWebBase("https://gitlab.example.com/group/proj.git"); got != "https://gitlab.example.com/group/proj" {
		t.Fatalf("got %q", got)
	}
}
