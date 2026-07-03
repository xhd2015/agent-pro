// Package dialog implements macOS human-in-the-loop dialogs via osascript.
package dialog

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	customizeButton  = "Customize"
	maxAlertButtons  = 3
	MaxPresetOptions = maxAlertButtons - 1 // preset --option count (Customize is always added)
)

// ErrDismissed indicates the user cancelled or closed the dialog.
var ErrDismissed = errors.New("dialog dismissed")

// AskRequest describes a two-step human checkpoint dialog.
type AskRequest struct {
	Title        string
	Message      string
	Options      []string
	AffirmOption string
	CancelOption string
}

// AskResult is the structured answer from Ask.
type AskResult struct {
	Answer   string
	Via      string // "button" | "free_text" | "dismissed"
	Affirmed *bool
}

// Escape turns arbitrary text into a safe AppleScript string literal fragment.
func Escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ParseOsascriptOutput extracts button and text from osascript stdout.
// display dialog often returns both fields on one line:
// "button returned:OK, text returned:hello"
func ParseOsascriptOutput(output string) (button, text string, err error) {
	button = extractOsascriptField(output, "button returned:")
	text = extractOsascriptField(output, "text returned:")
	return button, text, nil
}

func extractOsascriptField(output, prefix string) string {
	idx := strings.Index(output, prefix)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(output[idx+len(prefix):])
	if rest == "" {
		return ""
	}
	end := len(rest)
	for _, delim := range []string{", button returned:", ", text returned:", "\n", "\r"} {
		if pos := strings.Index(rest, delim); pos >= 0 && pos < end {
			end = pos
		}
	}
	return strings.TrimSpace(rest[:end])
}

func isDryRun() bool {
	return os.Getenv("DEBUG_WITH_USER_DRY_RUN") == "1"
}

// Ask presents the human checkpoint dialog and returns the user's answer.
func Ask(req AskRequest) (AskResult, error) {
	if req.CancelOption == "" {
		req.CancelOption = "Cancel"
	}
	if err := validateAskRequest(req); err != nil {
		return AskResult{}, err
	}
	if isDryRun() {
		return askDryRun(req)
	}
	if runtime.GOOS != "darwin" {
		return AskResult{}, fmt.Errorf(
			"human dialogs require macOS and osascript; on non-macOS set DEBUG_WITH_USER_DRY_RUN=1 for CI",
		)
	}
	return askOsascript(req)
}

func askDryRun(req AskRequest) (AskResult, error) {
	if os.Getenv("DEBUG_WITH_USER_DRY_RUN_DISMISSED") == "1" {
		return AskResult{Via: "dismissed"}, nil
	}
	button := strings.TrimSpace(os.Getenv("DEBUG_WITH_USER_DRY_RUN_BUTTON"))
	if button == "" {
		return AskResult{}, fmt.Errorf("dry-run requires DEBUG_WITH_USER_DRY_RUN_BUTTON")
	}
	if button == customizeButton {
		text := os.Getenv("DEBUG_WITH_USER_DRY_RUN_TEXT")
		if text == "" {
			return AskResult{}, fmt.Errorf("dry-run Customize requires DEBUG_WITH_USER_DRY_RUN_TEXT")
		}
		return AskResult{Answer: text, Via: "free_text"}, nil
	}
	affirmed := button == req.AffirmOption
	return AskResult{Answer: button, Via: "button", Affirmed: &affirmed}, nil
}

func askOsascript(req AskRequest) (AskResult, error) {
	buttons := buildAlertButtons(req)
	buttonSpec := make([]string, len(buttons))
	for i, b := range buttons {
		buttonSpec[i] = `"` + Escape(b) + `"`
	}
	defaultButton := req.AffirmOption
	if defaultButton == "" || !containsString(buttons, defaultButton) {
		if len(buttons) > 0 {
			defaultButton = buttons[0]
		}
	}
	// Do not set cancel button on preset options: macOS treats cancel-button
	// clicks as "User canceled" (-128) instead of returning the label.
	script := fmt.Sprintf(
		`display alert "%s" message "%s" as informational buttons {%s} default button "%s"`,
		Escape(req.Title),
		Escape(req.Message),
		strings.Join(buttonSpec, ", "),
		Escape(defaultButton),
	)

	output, err := runOsascript(script)
	if err != nil {
		if errors.Is(err, ErrDismissed) {
			return AskResult{Via: "dismissed"}, nil
		}
		return AskResult{}, err
	}
	button, _, err := ParseOsascriptOutput(output)
	if err != nil {
		return AskResult{}, err
	}
	if button == customizeButton {
		return askCustomizeText(req)
	}
	affirmed := button == req.AffirmOption
	return AskResult{Answer: button, Via: "button", Affirmed: &affirmed}, nil
}

func customizeDialogContent(req AskRequest) (title, body string) {
	title = strings.TrimSpace(req.Title)
	if title == "" {
		title = "Describe what happened"
	}
	body = strings.TrimSpace(req.Message)
	if body == "" {
		body = "Type what actually happened:"
	} else {
		body += "\n\nType what actually happened:"
	}
	return title, body
}

func askCustomizeText(req AskRequest) (AskResult, error) {
	title, body := customizeDialogContent(req)
	cancel := req.CancelOption
	if cancel == "" {
		cancel = "Cancel"
	}
	script := fmt.Sprintf(
		`display dialog "%s" default answer "" with title "%s" buttons {"OK", "%s"} default button "OK" cancel button "%s"`,
		Escape(body),
		Escape(title),
		Escape(cancel),
		Escape(cancel),
	)
	output, err := runOsascript(script)
	if err != nil {
		if errors.Is(err, ErrDismissed) {
			return AskResult{Via: "dismissed"}, nil
		}
		return AskResult{}, err
	}
	_, text, err := ParseOsascriptOutput(output)
	if err != nil {
		return AskResult{}, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return AskResult{Via: "dismissed"}, nil
	}
	return AskResult{Answer: text, Via: "free_text"}, nil
}

func validateAskRequest(req AskRequest) error {
	given := len(req.Options) + 1 // presets plus Customize
	if given > maxAlertButtons {
		return fmt.Errorf(
			"at most %d options allowed, you're giving %d options; consider breaking down the flow into multi-step confirmation",
			maxAlertButtons, given,
		)
	}
	return nil
}

// buildAlertButtons returns preset options plus Customize (caller must validate first).
func buildAlertButtons(req AskRequest) []string {
	return append(append([]string{}, req.Options...), customizeButton)
}

func classifyOsascriptFailure(output string) error {
	if strings.Contains(output, "User canceled") {
		return ErrDismissed
	}
	if strings.Contains(output, "execution error:") {
		return fmt.Errorf("osascript: %s", strings.TrimSpace(output))
	}
	return nil
}

func containsString(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func runOsascript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	output := string(out)
	trimmed := strings.TrimSpace(output)
	if err != nil {
		if classified := classifyOsascriptFailure(trimmed); classified != nil {
			return output, classified
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return output, ErrDismissed
		}
		return output, fmt.Errorf("osascript failed: %w: %s", err, trimmed)
	}
	return output, nil
}