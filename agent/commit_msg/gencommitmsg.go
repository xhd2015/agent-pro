package commit_msg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/exec/tool_exec"
	"github.com/xhd2015/agent-pro/agent/git_runner"
	"github.com/xhd2015/agent-pro/agent/opencode/models"
	"github.com/xhd2015/agent-pro/agent/opencode/run"
	"github.com/xhd2015/dot-pkgs/go-pkgs/file/detect"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/submodule"
	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/gitops/gitwrite"
	"github.com/xhd2015/less-gen/flags"
)

var genCommitMsgHelp = `Usage: gen-commit-msg [options]

Generate a commit message for the currently staged changes using AI.
Logs are printed to stderr; the resulting commit message is printed to stdout.

Options:
  --dir DIR    Git directory to use (defaults to current directory)
  --model MODEL
              Model to use for generation
  --agent-runner RUNNER
              Agent runner to use (opencode|commandcode, default: opencode)
  --agent-runner-binary PATH
              Override the agent runner executable path
  --commit     Run git commit with the generated message after printing it
  --no-verify  Skip git commit hooks (requires --commit)
  --dry-run    Pure plan: inspect staged set, print mock message; no agent, no unstage, no commit
  -h, --help   Show this help message
`

func RunGenCommitMsg(args []string) error {
	var dir string
	var model string
	var agentRunner string
	var agentRunnerBinary string
	var commit bool
	var noVerify bool
	var dryRun bool
	_, err := flags.
		String("--dir", &dir).
		String("--model", &model).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		Bool("--commit", &commit).
		Bool("--no-verify", &noVerify).
		Bool("--dry-run", &dryRun).
		Help("-h,--help", genCommitMsgHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if noVerify && !commit {
		return fmt.Errorf("--no-verify requires --commit")
	}

	if dir == "" {
		dir, _ = os.Getwd()
	}
	if agentRunner == "" {
		agentRunner = "opencode"
	}
	if agentRunner != "opencode" && agentRunner != "commandcode" {
		return fmt.Errorf("unsupported agent runner: %s (supported: opencode, commandcode)", agentRunner)
	}

	inside, err := git.IsInsideGit(dir)
	if err != nil {
		return err
	}
	if !inside {
		return fmt.Errorf("not a git repository: %s", dir)
	}

	if dryRun {
		return runGenCommitMsgDryRun(dir, commit, noVerify)
	}

	msg, err := Generate(dir, GenerateOptions{
		Model:             model,
		AgentRunner:       agentRunner,
		AgentRunnerBinary: agentRunnerBinary,
		Logger:            &stderrLogger{},
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\n--- Generated Commit Message ---\n")
	fmt.Println(msg)

	quotedMsg := shellQuote(msg)
	commitCmd := "git commit -m " + quotedMsg
	if noVerify {
		commitCmd += " --no-verify"
	}
	fmt.Fprintf(os.Stderr, "\nRun:\n  %s\n", commitCmd)

	if commit {
		fmt.Fprintf(os.Stderr, "\nRunning git commit...\n")
		output, err := git_runner.CommitWithRetry(dir, msg, 5, noVerify)
		if len(output) > 0 {
			fmt.Fprint(os.Stderr, string(output))
		}
		if err != nil {
			return fmt.Errorf("git commit failed: %w", err)
		}
	}

	return nil
}

// runGenCommitMsgDryRun implements pure-plan --dry-run: inspect staged set, print
// mock message B, plan would-unstage / would-commit on stderr, never call agent
// or mutate the index/HEAD.
func runGenCommitMsgDryRun(dir string, commit, noVerify bool) error {
	stagedFiles, err := git.GetStagedFiles(dir)
	if err != nil {
		return fmt.Errorf("failed to list staged files: %w", err)
	}
	if len(stagedFiles) == 0 {
		return fmt.Errorf("no staged changes to generate commit message for")
	}

	// Plan binary/submodule unstage without mutating the index.
	binaries, subModuleDirs, err := detectUnstageCandidates(dir, stagedFiles)
	if err != nil {
		return fmt.Errorf("auto unstage failed: %w", err)
	}
	for _, b := range binaries {
		if b.desc != "" {
			fmt.Fprintf(os.Stderr, "would: unstage %s (%s)\n", b.path, b.desc)
		} else {
			fmt.Fprintf(os.Stderr, "would: unstage %s\n", b.path)
		}
	}
	for _, sm := range subModuleDirs {
		fmt.Fprintf(os.Stderr, "would: unstage %s/\n", sm)
	}

	n := len(stagedFiles)
	fmt.Printf("dry-run: would generate commit message for %d staged file(s)\n", n)

	if commit {
		// Plan commit using the mock message as the planned -m payload.
		mockMsg := fmt.Sprintf("dry-run: would generate commit message for %d staged file(s)", n)
		commitCmd := "git commit -m " + shellQuote(mockMsg)
		if noVerify {
			commitCmd += " --no-verify"
		}
		fmt.Fprintf(os.Stderr, "would: %s\n", commitCmd)
	}

	return nil
}

type stderrLogger struct{}

func (l *stderrLogger) Log(msg string)   { fmt.Fprintf(os.Stderr, "%s\n", msg) }
func (l *stderrLogger) Error(msg string) { fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg) }

func shellQuote(s string) string {
	if !strings.ContainsAny(s, "'\"\\$ !`\n\r\t") {
		return "'" + s + "'"
	}
	return "$'" + strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	).Replace(s) + "'"
}

type Logger interface {
	Log(msg string)
	Error(msg string)
}

type GenerateOptions struct {
	Model             string
	AgentRunner       string
	AgentRunnerBinary string
	AgentEnv          map[string]string
	Logger            Logger
}

func Generate(dir string, options GenerateOptions) (string, error) {
	logger := options.Logger
	optionModel := options.Model

	if err := detectAndUnstage(dir, logger); err != nil {
		return "", fmt.Errorf("auto unstage failed: %w", err)
	}

	logger.Log("$ git diff --cached")
	diff, err := git.DiffCached(dir)
	if err != nil {
		var pe *git.DiffCachedParseError
		if errors.As(err, &pe) {
			logger.Log(fmt.Sprintf("DiffCached parse error (raw length: %d chars): %v", len(pe.Raw), err))
		}
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	if diff == nil {
		return "", fmt.Errorf("no staged changes to generate commit message for")
	}

	fileCount := diff.FileCount()
	stagedDiff := diff.UnifiedTruncated(24)

	logger.Log(fmt.Sprintf("Staged files: %d, Diff length: %d chars", fileCount, len(diff.Raw)))
	logger.Log(fmt.Sprintf("Truncated diff length: %d chars", len(stagedDiff)))
	logger.Log("Passing diff to agent...")

	commitPrompt := fmt.Sprintf(`Generate a brief git commit message (1 line title, max 50 characters, plus a short description if needed) for the following staged changes (git diff). Focus on what changed and why.

Git diff:
%s

Respond with ONLY a JSON object in this exact format (no other text):
{"title": "<short title>", "description": "<optional short description>"}`, stagedDiff)

	runner := options.AgentRunner
	if runner == "" {
		runner = "opencode"
	}

	var rawText string
	switch runner {
	case "commandcode":
		out, err := runCommandCodeAgent(dir, options, commitPrompt, fileCount, len(stagedDiff))
		if err != nil {
			return "", err
		}
		rawText = strings.TrimSpace(out)
		if rawText == "" {
			return "", fmt.Errorf("failed to parse commit message from commandcode output")
		}
	case "opencode":
		logger.Log("$ opencode models")
		freeModels, preferredModel, err := models.ListFree()
		if err != nil {
			logger.Log(fmt.Sprintf("Warning: Could not get free models: %v", err))
		} else {
			logger.Log(fmt.Sprintf("Free models: %s", strings.Join(freeModels, ", ")))
			if preferredModel != "" && optionModel == "" {
				logger.Log(fmt.Sprintf("Free model suggestion: %s", preferredModel))
			}
		}

		actualModel := preferredModel
		if optionModel != "" {
			actualModel = optionModel
		}
		logger.Log(fmt.Sprintf("Using model: %s", actualModel))

		logger.Log(fmt.Sprintf("$ opencode run  [prompt: Generate brief git commit message for %d staged file(s), %d chars]", fileCount, len(stagedDiff)))
		logger.Log("Running agent...")

		ctx := context.Background()
		output, sessionID, err := run.Run(ctx, run.Options{
			Dir:       dir,
			Model:     actualModel,
			Prompt:    commitPrompt,
			Logger:    logger,
			AgentPath: options.AgentRunnerBinary,
			Env:       options.AgentEnv,
		})
		_ = sessionID
		if err != nil {
			return "", fmt.Errorf("agent failed: %w", err)
		}

		rawText = parseOpencodeJSONOutput(output)
		if rawText == "" {
			rawText = strings.TrimSpace(output)
		}
		if rawText == "" {
			return "", fmt.Errorf("failed to parse commit message from opencode output")
		}
	default:
		return "", fmt.Errorf("unsupported agent runner: %s (supported: opencode, commandcode)", runner)
	}

	// Post-parse sanitize choke point: strip anti-patterns; hard-fail if unusable.
	sanitized, err := SanitizeOrError(rawText)
	if err != nil {
		return "", err
	}
	commitMessage := sanitized.format()
	if commitMessage == "" {
		return "", fmt.Errorf("failed to extract commit message from agent response")
	}

	return commitMessage, nil
}

// runCommandCodeAgent invokes Command Code (`cmd` on PATH, or AgentRunnerBinary override).
// Argv: -p <prompt> --skip-onboarding --yolo --max-turns 1 [-m <model>]
// Full stdout is returned (no opencode NDJSON parse).
func runCommandCodeAgent(dir string, options GenerateOptions, commitPrompt string, fileCount, stagedDiffLen int) (string, error) {
	logger := options.Logger

	// Skip models.ListFree for commandcode; only honor an explicit --model.
	if options.Model != "" {
		logger.Log(fmt.Sprintf("Using model: %s", options.Model))
	}

	args := []string{"-p", commitPrompt, "--skip-onboarding", "--yolo", "--max-turns", "1"}
	if options.Model != "" {
		args = append(args, "-m", options.Model)
	}

	binLabel := "cmd"
	if options.AgentRunnerBinary != "" {
		binLabel = filepath.Base(options.AgentRunnerBinary)
	}
	logger.Log(fmt.Sprintf("$ %s -p …  [prompt: Generate brief git commit message for %d staged file(s), %d chars]", binLabel, fileCount, stagedDiffLen))
	logger.Log("Running agent...")

	cmd, err := tool_exec.New("cmd", args, &tool_exec.Options{
		Dir:        dir,
		CustomPath: options.AgentRunnerBinary,
		Env:        options.AgentEnv,
	})
	if err != nil {
		return "", fmt.Errorf("agent failed: %w", err)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if se := strings.TrimSpace(stderr.String()); se != "" {
			logger.Log(se)
		}
		return "", fmt.Errorf("agent failed: %w", err)
	}
	if se := strings.TrimSpace(stderr.String()); se != "" {
		logger.Log(se)
	}
	return stdout.String(), nil
}

type unstagedItem struct {
	path string
	desc string
}

// detectUnstageCandidates finds staged binaries and submodule paths that would
// be auto-unstaged. stagedFiles may be nil to list ACMRT-filtered staged paths.
// Paths are evaluated from the repo root (same chdir semantics as detectAndUnstage).
func detectUnstageCandidates(dir string, stagedFiles []string) (binaries []unstagedItem, subModuleDirs []string, err error) {
	dir = filepath.Clean(dir)

	repoRoot, err := git.ShowToplevel(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve git toplevel: %w", err)
	}
	if repoRoot == "" {
		return nil, nil, fmt.Errorf("empty git toplevel")
	}

	origDir, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("getwd: %w", err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		return nil, nil, fmt.Errorf("chdir %s: %w", repoRoot, err)
	}
	defer os.Chdir(origDir)

	if stagedFiles == nil {
		stagedFiles, err = git.GetStagedFiles(dir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list staged files: %w", err)
		}
	}
	if len(stagedFiles) == 0 {
		return nil, nil, nil
	}

	subModuleDirs = submodule.DetectSubModules(stagedFiles)

	isInSubmodule := func(f string) bool {
		for _, smDir := range subModuleDirs {
			if strings.HasPrefix(f, smDir+string(filepath.Separator)) || f == smDir {
				return true
			}
		}
		return false
	}

	for _, f := range stagedFiles {
		if isInSubmodule(f) {
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, nil, fmt.Errorf("stat %s: %w", f, err)
		}
		if info.IsDir() {
			continue
		}
		desc, isBin, err := detect.DetectFileType(f)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, nil, fmt.Errorf("detect %s: %w", f, err)
		}
		if isBin {
			binaries = append(binaries, unstagedItem{path: f, desc: desc})
		}
	}

	return binaries, subModuleDirs, nil
}

func detectAndUnstage(dir string, logger Logger) error {
	binaries, subModuleDirs, err := detectUnstageCandidates(dir, nil)
	if err != nil {
		return err
	}

	// Rebuild toUnstage list (binaries + any staged path under a submodule).
	// Re-list ACMRT staged files for submodule file paths.
	stagedFiles, err := git.GetStagedFiles(dir)
	if err != nil {
		return fmt.Errorf("failed to list staged files: %w", err)
	}

	isInSubmodule := func(f string) bool {
		for _, smDir := range subModuleDirs {
			if strings.HasPrefix(f, smDir+string(filepath.Separator)) || f == smDir {
				return true
			}
		}
		return false
	}

	var toUnstage []string
	for _, b := range binaries {
		toUnstage = append(toUnstage, b.path)
	}
	for _, f := range stagedFiles {
		if isInSubmodule(f) {
			toUnstage = append(toUnstage, f)
		}
	}

	if len(toUnstage) == 0 {
		return nil
	}

	if len(binaries) > 0 {
		fmt.Fprintf(os.Stderr, "\nAuto-unstaged binary files:\n")
		for _, b := range binaries {
			if b.desc != "" {
				fmt.Fprintf(os.Stderr, "  %s (%s)\n", b.path, b.desc)
			} else {
				fmt.Fprintf(os.Stderr, "  %s\n", b.path)
			}
		}
	}
	if len(subModuleDirs) > 0 {
		fmt.Fprintf(os.Stderr, "\nAuto-unstaged submodule directories:\n")
		for _, sm := range subModuleDirs {
			fmt.Fprintf(os.Stderr, "  %s/\n", sm)
		}
	}

	if len(toUnstage) > 0 {
		fmt.Fprintf(os.Stderr, "\nTo include these files back: use `git add <file> && git commit --amend --no-edit`\n")
	}

	logger.Log("Unstaging binary/submodule entries...")
	if err := gitwrite.RestoreStaged(dir, toUnstage...); err != nil {
		return err
	}

	return nil
}

type CommitMsg struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (m CommitMsg) format() string {
	if m.Title == "" {
		return ""
	}
	if m.Description == "" {
		return m.Title
	}
	return m.Title + "\n\n" + m.Description
}

func parseCommitMsgFromText(text string) string {
	jsonText := extractJSONFromText(text)
	if jsonText != "" {
		var msg CommitMsg
		if err := json.Unmarshal([]byte(jsonText), &msg); err == nil && msg.Title != "" {
			return msg.format()
		}
	}
	return stripCommitHeaders(text)
}

func extractJSONFromText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if text[0] == '{' && text[len(text)-1] == '}' {
		if json.Valid([]byte(text)) {
			return text
		}
	}
	if idx := strings.Index(text, "```"); idx >= 0 {
		rest := text[idx+3:]
		rest = strings.TrimPrefix(rest, "json")
		rest = strings.TrimSpace(rest)
		if endIdx := strings.Index(rest, "```"); endIdx >= 0 {
			candidate := strings.TrimSpace(rest[:endIdx])
			if json.Valid([]byte(candidate)) {
				return candidate
			}
		}
	}
	firstBrace := strings.Index(text, "{")
	lastBrace := strings.LastIndex(text, "}")
	if firstBrace >= 0 && lastBrace > firstBrace {
		candidate := text[firstBrace : lastBrace+1]
		if json.Valid([]byte(candidate)) {
			return candidate
		}
	}
	return ""
}

func stripCommitHeaders(msg string) string {
	lines := strings.Split(msg, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, prefix := range []string{"Title:", "title:", "Description:", "description:"} {
			if strings.HasPrefix(trimmed, prefix) {
				trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
				break
			}
		}
		result = append(result, trimmed)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

func parseOpencodeJSONOutput(output string) string {
	lines := strings.Split(output, "\n")
	var currentStepText strings.Builder
	var lastStopText string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		eventType, _ := event["type"].(string)
		part, _ := event["part"].(map[string]interface{})
		if part == nil {
			continue
		}

		switch eventType {
		case "step_start":
			currentStepText.Reset()
		case "text":
			if text, ok := part["text"].(string); ok {
				currentStepText.WriteString(text)
			}
		case "step_finish":
			text := currentStepText.String()
			if text != "" {
				lastStopText = text
			}
		}
	}

	return strings.TrimSpace(lastStopText)
}
