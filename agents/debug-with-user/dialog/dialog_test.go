package dialog

import (
	"errors"
	"testing"
)

func TestBuildAlertButtons_twoOptionsPlusCustomize(t *testing.T) {
	buttons := buildAlertButtons(AskRequest{
		Options:      []string{"Looks good", "Something wrong"},
		AffirmOption: "Looks good",
		CancelOption: "Cancel",
	})
	if len(buttons) != 3 {
		t.Fatalf("buttons = %v, want 3", buttons)
	}
	if buttons[2] != customizeButton {
		t.Fatalf("buttons = %v, want Customize last", buttons)
	}
}

func TestParseOsascriptOutput_combinedLine(t *testing.T) {
	button, text, err := ParseOsascriptOutput("button returned:OK, text returned:hello world\n")
	if err != nil {
		t.Fatal(err)
	}
	if button != "OK" {
		t.Fatalf("button = %q, want OK", button)
	}
	if text != "hello world" {
		t.Fatalf("text = %q, want hello world", text)
	}
}

func TestParseOsascriptOutput_separateLines(t *testing.T) {
	button, text, err := ParseOsascriptOutput("button returned:Customize\n")
	if err != nil {
		t.Fatal(err)
	}
	if button != "Customize" {
		t.Fatalf("button = %q", button)
	}
	if text != "" {
		t.Fatalf("text = %q, want empty", text)
	}

	_, text, err = ParseOsascriptOutput("text returned:VS Code opened but wrong workspace\n")
	if err != nil {
		t.Fatal(err)
	}
	if text != "VS Code opened but wrong workspace" {
		t.Fatalf("text = %q", text)
	}
}

func TestCustomizeDialogContent_includesTitleAndMessage(t *testing.T) {
	title, body := customizeDialogContent(AskRequest{
		Title:   "Smoke test",
		Message: "Pick an option or Customize to type a report.",
	})
	if title != "Smoke test" {
		t.Fatalf("title = %q", title)
	}
	if body != "Pick an option or Customize to type a report.\n\nType what actually happened:" {
		t.Fatalf("body = %q", body)
	}
}

func TestCustomizeDialogContent_defaultsWithoutTitle(t *testing.T) {
	title, body := customizeDialogContent(AskRequest{Message: "Did it work?"})
	if title != "Describe what happened" {
		t.Fatalf("title = %q", title)
	}
	if body != "Did it work?\n\nType what actually happened:" {
		t.Fatalf("body = %q", body)
	}
}

func TestClassifyOsascriptFailure_userCanceled(t *testing.T) {
	err := classifyOsascriptFailure("0:146: execution error: User canceled. (-128)")
	if !errors.Is(err, ErrDismissed) {
		t.Fatalf("err = %v, want ErrDismissed", err)
	}
}

func TestValidateAskRequest_tooManyPresets(t *testing.T) {
	err := validateAskRequest(AskRequest{
		Options: []string{"Looks good", "Something wrong2", "Something wrong3"},
	})
	if err == nil {
		t.Fatal("expected error for too many options")
	}
}

func TestAsk_rejectsTooManyPresets(t *testing.T) {
	t.Setenv("DEBUG_WITH_USER_DRY_RUN", "1")
	t.Setenv("DEBUG_WITH_USER_DRY_RUN_BUTTON", "Looks good")

	_, err := Ask(AskRequest{
		Options: []string{"Looks good", "Something wrong2", "Something wrong3"},
	})
	if err == nil {
		t.Fatal("expected error for too many options")
	}
}