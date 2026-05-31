package commit_msg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type testLogger struct {
	logs []string
}

func (l *testLogger) Log(msg string)   { l.logs = append(l.logs, msg) }
func (l *testLogger) Error(msg string) { l.logs = append(l.logs, "ERROR: "+msg) }

func TestDetectAndUnstage_NoStagedFiles(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDetectAndUnstage_BinaryFile(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	writeFile(t, "image.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00})
	writeFile(t, "readme.md", []byte("hello\n"))
	mustRun(t, "git", "add", "image.png", "readme.md")

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	staged := getStagedFiles(t)
	if containsFile(staged, "image.png") {
		t.Errorf("image.png should have been unstaged, but it's still staged")
	}
	if !containsFile(staged, "readme.md") {
		t.Errorf("readme.md should still be staged, but it's not")
	}
}

func TestDetectAndUnstage_SubmoduleDir(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	smDir := filepath.Join(dir, "vendor", "libfoo")
	mustRun(t, "mkdir", "-p", filepath.Join(smDir, "src"))
	writeFile(t, "vendor/libfoo/src/main.c", []byte("int main() { return 0; }\n"))
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	mustRun(t, "git", "add", "vendor/libfoo/src/main.c")

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	staged := getStagedFiles(t)
	if containsFile(staged, "vendor/libfoo/src/main.c") {
		t.Errorf("submodule file should have been unstaged, but it's still staged")
	}
}

func TestDetectAndUnstage_SubmoduleWorktree(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	smDir := filepath.Join(dir, "vendor", "libbar")
	mustRun(t, "mkdir", "-p", smDir)
	writeFile(t, "vendor/libbar/main.go", []byte("package main\n"))
	writeFile(t, "vendor/libbar/.git", []byte("gitdir: ../../.git/modules/vendor/libbar"))
	mustRun(t, "git", "add", "vendor/libbar/main.go")

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	staged := getStagedFiles(t)
	if containsFile(staged, "vendor/libbar/main.go") {
		t.Errorf("worktree submodule file should have been unstaged")
	}
}

func TestDetectAndUnstage_BinaryAndSubmodule(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	writeFile(t, "icon.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00})
	writeFile(t, "README.md", []byte("hello\n"))

	smDir := filepath.Join(dir, "vendor", "libfoo")
	mustRun(t, "mkdir", "-p", filepath.Join(smDir, "src"))
	writeFile(t, "vendor/libfoo/config.h", []byte("#define VERSION 1\n"))
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)

	mustRun(t, "git", "add", "icon.png", "README.md", "vendor/libfoo/config.h")

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	staged := getStagedFiles(t)
	if containsFile(staged, "icon.png") {
		t.Errorf("icon.png should have been unstaged")
	}
	if containsFile(staged, "vendor/libfoo/config.h") {
		t.Errorf("submodule file should have been unstaged")
	}
	if !containsFile(staged, "README.md") {
		t.Errorf("README.md should still be staged")
	}
}

func TestDetectAndUnstage_NoOffendingFiles(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	writeFile(t, "README.md", []byte("hello\n"))
	writeFile(t, "src/main.go", []byte("package main\n"))
	mustRun(t, "git", "add", "README.md", "src/main.go")

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	staged := getStagedFiles(t)
	if !containsFile(staged, "README.md") {
		t.Errorf("README.md should still be staged")
	}
	if !containsFile(staged, "src/main.go") {
		t.Errorf("src/main.go should still be staged")
	}
}

func TestDetectAndUnstage_SubmoduleFileIsAlsoBinary(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	smDir := filepath.Join(dir, "vendor", "libfoo")
	mustRun(t, "mkdir", "-p", smDir)
	writeFile(t, "vendor/libfoo/binary.dat", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00})
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	mustRun(t, "git", "add", "vendor/libfoo/binary.dat")

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	staged := getStagedFiles(t)
	if containsFile(staged, "vendor/libfoo/binary.dat") {
		t.Errorf("submodule file should have been unstaged (submodule takes priority over binary)")
	}
}

func TestDetectAndUnstage_OutputFormat(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	writeFile(t, "icon.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00})

	smDir := filepath.Join(dir, "vendor", "libfoo")
	mustRun(t, "mkdir", "-p", smDir)
	writeFile(t, "vendor/libfoo/README.md", []byte("hello\n"))
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)

	mustRun(t, "git", "add", "icon.png", "vendor/libfoo/README.md")

	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := &testLogger{}
	err := detectAndUnstage(dir, logger)

	w.Close()
	var buf [8192]byte
	n, _ := r.Read(buf[:])
	os.Stderr = origStderr
	output := string(buf[:n])

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Auto-unstaged binary files:") {
		t.Errorf("output should contain binary section header, got:\n%s", output)
	}
	if !strings.Contains(output, "icon.png") {
		t.Errorf("output should contain binary file name, got:\n%s", output)
	}
	if !strings.Contains(output, "Auto-unstaged submodule directories:") {
		t.Errorf("output should contain submodule section header, got:\n%s", output)
	}
	if !strings.Contains(output, "vendor/libfoo/") {
		t.Errorf("output should contain submodule directory, got:\n%s", output)
	}
	if !strings.Contains(output, "To include these files back:") {
		t.Errorf("output should contain amend hint, got:\n%s", output)
	}
	if !strings.Contains(output, "git commit --amend --no-edit") {
		t.Errorf("output should contain --amend flag, got:\n%s", output)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gencommitmsg-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %s: %v", string(out), err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	cmd.Run()

	_ = os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0644)
	cmd = exec.Command("git", "add", ".gitkeep")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = dir
	cmd.Run()

	return dir
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRun(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed: %s: %v", name, args, string(out), err)
	}
}

func getStagedFiles(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACMRT")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git diff --cached --name-only failed: %v", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func containsFile(files []string, name string) bool {
	for _, f := range files {
		if f == name {
			return true
		}
	}
	return false
}

func TestCommitMsgFormat(t *testing.T) {
	tests := []struct {
		name string
		msg  CommitMsg
		want string
	}{
		{
			name: "title only",
			msg:  CommitMsg{Title: "Add feature"},
			want: "Add feature",
		},
		{
			name: "title and description",
			msg:  CommitMsg{Title: "Add feature", Description: "Implement feature X"},
			want: "Add feature\n\nImplement feature X",
		},
		{
			name: "empty title with description",
			msg:  CommitMsg{Title: "", Description: "desc"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.format()
			if got != tt.want {
				t.Errorf("format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractJSONFromText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "clean JSON",
			text: `{"title": "Add feature", "description": "details"}`,
			want: `{"title": "Add feature", "description": "details"}`,
		},
		{
			name: "JSON with leading whitespace",
			text: `  {"title": "Fix bug"}  `,
			want: `{"title": "Fix bug"}`,
		},
		{
			name: "markdown code fence with json tag",
			text: "```json\n{\"title\": \"Add feature\"}\n```",
			want: `{"title": "Add feature"}`,
		},
		{
			name: "markdown code fence without language tag",
			text: "```\n{\"title\": \"Fix bug\"}\n```",
			want: `{"title": "Fix bug"}`,
		},
		{
			name: "JSON with extra text before and after",
			text: `Here is the commit message: {"title": "Add feature", "description": "details"} hope that works`,
			want: `{"title": "Add feature", "description": "details"}`,
		},
		{
			name: "no JSON in text",
			text: "just some text without braces",
			want: "",
		},
		{
			name: "empty text",
			text: "",
			want: "",
		},
		{
			name: "JSON without closing brace",
			text: `{"title": "Add feature"`,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONFromText(tt.text)
			if got != tt.want {
				t.Errorf("extractJSONFromText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCommitMsgFromText_JSON(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "clean JSON with title and description",
			text: `{"title": "Add feature", "description": "Implement feature X"}`,
			want: "Add feature\n\nImplement feature X",
		},
		{
			name: "clean JSON title only",
			text: `{"title": "Fix bug"}`,
			want: "Fix bug",
		},
		{
			name: "JSON with description empty",
			text: `{"title": "Refactor code", "description": ""}`,
			want: "Refactor code",
		},
		{
			name: "JSON in markdown fence",
			text: "```json\n{\"title\": \"Update README\", \"description\": \"Add setup instructions\"}\n```",
			want: "Update README\n\nAdd setup instructions",
		},
		{
			name: "JSON with extra text",
			text: `ok here: {"title": "Optimize query", "description": "Add index"}`,
			want: "Optimize query\n\nAdd index",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommitMsgFromText(tt.text)
			if got != tt.want {
				t.Errorf("parseCommitMsgFromText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCommitMsgFromText_Fallback(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "legacy Title: Description: format",
			text: "Title: Add feature\nDescription: Implement feature X",
			want: "Add feature\nImplement feature X",
		},
		{
			name: "legacy lowercase prefixes",
			text: "title: Fix bug\ndescription: Critical fix",
			want: "Fix bug\nCritical fix",
		},
		{
			name: "mixed legacy format with indentation",
			text: "  Title:  Add login  \n  Description:  OAuth2 based  ",
			want: "Add login\nOAuth2 based",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommitMsgFromText(tt.text)
			if got != tt.want {
				t.Errorf("parseCommitMsgFromText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseOpencodeJSONOutput_JSONAgentResponse(t *testing.T) {
	ndjson := `{"type":"step_start","part":{}}
{"type":"text","part":{"text":"{\"title\": \"Add feature\", \"description\": \"Implement feature X\"}"}}
{"type":"step_finish","part":{}}`

	got := parseOpencodeJSONOutput(ndjson)
	want := `{"title": "Add feature", "description": "Implement feature X"}`
	if got != want {
		t.Errorf("parseOpencodeJSONOutput() = %q, want %q", got, want)
	}
}

func TestParseOpencodeJSONOutput_MultipleSteps(t *testing.T) {
	ndjson := `{"type":"step_start","part":{}}
{"type":"text","part":{"text":"first step"}}
{"type":"step_finish","part":{}}
{"type":"step_start","part":{}}
{"type":"text","part":{"text":"{\"title\": \"Final msg\"}"}}
{"type":"step_finish","part":{}}`

	got := parseOpencodeJSONOutput(ndjson)
	want := `{"title": "Final msg"}`
	if got != want {
		t.Errorf("parseOpencodeJSONOutput() = %q, want %q", got, want)
	}
}

func TestParseOpencodeJSONOutput_EmptyStep(t *testing.T) {
	ndjson := `{"type":"step_start","part":{}}
{"type":"step_finish","part":{}}
{"type":"step_start","part":{}}
{"type":"text","part":{"text":"{\"title\": \"Fix bug\"}"}}
{"type":"step_finish","part":{}}`

	got := parseOpencodeJSONOutput(ndjson)
	want := `{"title": "Fix bug"}`
	if got != want {
		t.Errorf("parseOpencodeJSONOutput() = %q, want %q", got, want)
	}
}

func TestParseOpencodeJSONOutput_EmptyOutput(t *testing.T) {
	got := parseOpencodeJSONOutput("")
	if got != "" {
		t.Errorf("parseOpencodeJSONOutput() = %q, want empty", got)
	}
}

func TestParseCommitMsgFromText_EmptyJSONTitle(t *testing.T) {
	text := `{"title": "", "description": "just desc"}`
	got := parseCommitMsgFromText(text)
	// Should fallback to stripCommitHeaders, which won't find Title:/Description: prefixes
	// So the whole text goes through stripCommitHeaders which strips Title:/Description: prefixes
	// and joins the remaining lines
	if got == "" {
		t.Errorf("parseCommitMsgFromText() should have fallback output, got empty")
	}
}
