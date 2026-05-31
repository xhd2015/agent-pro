package commit_msg

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/git_runner"
	"github.com/xhd2015/agent-pro/agent/opencode/models"
	"github.com/xhd2015/agent-pro/agent/opencode/run"
	"github.com/xhd2015/dot-pkgs/go-pkgs/file/detect"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/submodule"
	"github.com/xhd2015/less-gen/flags"
)

var genCommitMsgHelp = `Usage: gen-commit-msg [options]

Generate a commit message for the currently staged changes using AI.
Logs are printed to stderr; the resulting commit message is printed to stdout.

Options:
  --dir DIR    Git directory to use (defaults to current directory)
  --model MODEL
              Model to use for generation
  --commit     Run git commit with the generated message after printing it
  -h, --help   Show this help message
`

func RunGenCommitMsg(args []string) error {
	var dir string
	var model string
	var commit bool
	_, err := flags.
		String("--dir", &dir).
		String("--model", &model).
		Bool("--commit", &commit).
		Help("-h,--help", genCommitMsgHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if dir == "" {
		dir, _ = os.Getwd()
	}

	msg, err := Generate(dir, GenerateOptions{
		Model:  model,
		Logger: &stderrLogger{},
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\n--- Generated Commit Message ---\n")
	fmt.Println(msg)

	quotedMsg := shellQuote(msg)
	fmt.Fprintf(os.Stderr, "\nRun:\n  git commit -m %s\n", quotedMsg)

	if commit {
		fmt.Fprintf(os.Stderr, "\nRunning git commit...\n")
		output, err := git_runner.Commit(msg).Dir(dir).Run()
		if len(output) > 0 {
			fmt.Fprint(os.Stderr, string(output))
		}
		if err != nil {
			return fmt.Errorf("git commit failed: %w", err)
		}
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
	Model  string
	Logger Logger
}

func Generate(dir string, options GenerateOptions) (string, error) {
	logger := options.Logger
	optionModel := options.Model

	if err := detectAndUnstage(dir, logger); err != nil {
		return "", fmt.Errorf("auto unstage failed: %w", err)
	}

	logger.Log("$ git diff --cached")
	stagedDiffOutput, err := git_runner.DiffCached().Dir(dir).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	stagedDiff := string(stagedDiffOutput)
	if stagedDiff == "" {
		return "", fmt.Errorf("no staged changes to generate commit message for")
	}

	fileCount := strings.Count(stagedDiff, "diff --git")
	if fileCount == 0 && len(stagedDiff) > 0 {
		fileCount = 1
	}

	logger.Log(fmt.Sprintf("Staged files: %d, Diff length: %d chars", fileCount, len(stagedDiff)))

	const maxLinesPerFile = 24
	stagedDiff = truncateDiffPerFile(stagedDiff, maxLinesPerFile)

	logger.Log(fmt.Sprintf("Truncated diff length: %d chars", len(stagedDiff)))
	logger.Log("Passing diff to agent...")

	commitPrompt := fmt.Sprintf(`Generate a brief git commit message (1 line title, max 50 characters, plus a short description if needed) for the following staged changes (git diff). Focus on what changed and why.

Git diff:
%s

Respond with ONLY a JSON object in this exact format (no other text):
{"title": "<short title>", "description": "<optional short description>"}`, stagedDiff)

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

	output, err := run.Run(run.Options{
		Dir:    dir,
		Model:  actualModel,
		Prompt: commitPrompt,
		Logger: logger,
	})
	if err != nil {
		return "", fmt.Errorf("agent failed: %w", err)
	}

	rawText := parseOpencodeJSONOutput(output)
	if rawText == "" {
		return "", fmt.Errorf("failed to parse commit message from opencode output")
	}

	commitMessage := parseCommitMsgFromText(rawText)
	if commitMessage == "" {
		return "", fmt.Errorf("failed to extract commit message from agent response")
	}

	return commitMessage, nil
}

type unstagedItem struct {
	path string
	desc string
}

func detectAndUnstage(dir string, logger Logger) error {
	origDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("chdir %s: %w", dir, err)
	}
	defer os.Chdir(origDir)

	output, err := git_runner.NewCommand("diff", "--cached", "--name-only", "--diff-filter=ACMRT", "--").Output()
	if err != nil {
		return fmt.Errorf("failed to list staged files: %w", err)
	}

	var stagedFiles []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			stagedFiles = append(stagedFiles, name)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan staged files: %w", err)
	}
	if len(stagedFiles) == 0 {
		return nil
	}

	subModuleDirs := submodule.DetectSubModules(stagedFiles)

	isInSubmodule := func(f string) bool {
		for _, smDir := range subModuleDirs {
			if strings.HasPrefix(f, smDir+string(filepath.Separator)) || f == smDir {
				return true
			}
		}
		return false
	}

	var binaries []unstagedItem
	for _, f := range stagedFiles {
		if isInSubmodule(f) {
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("stat %s: %w", f, err)
		}
		if info.IsDir() {
			continue
		}
		desc, isBin, err := detect.DetectFileType(f)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("detect %s: %w", f, err)
		}
		if isBin {
			binaries = append(binaries, unstagedItem{path: f, desc: desc})
		}
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
	if err := unstageFiles(toUnstage); err != nil {
		return err
	}

	return nil
}

func unstageFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"restore", "--staged", "--"}, files...)
	output, err := git_runner.NewCommand(args...).Run()
	if err != nil {
		return fmt.Errorf("git restore --staged failed: %s: %w", string(output), err)
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

func truncateDiffPerFile(diff string, maxLines int) string {
	parts := strings.Split("\n"+diff, "\ndiff --git ")
	var b strings.Builder
	first := true
	for _, part := range parts {
		if part == "" {
			continue
		}
		section := "diff --git " + part
		if !first {
			b.WriteByte('\n')
		}
		first = false

		lines := strings.Split(strings.TrimRight(section, "\n"), "\n")
		if len(lines) <= maxLines {
			b.WriteString(strings.Join(lines, "\n"))
		} else {
			b.WriteString(strings.Join(lines[:maxLines], "\n"))
			b.WriteString(fmt.Sprintf("\n...(%d more lines omitted)", len(lines)-maxLines))
		}
	}
	return b.String()
}
