package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agents/debug-with-user/dialog"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/skills/install"
)

//go:embed SKILL.md
var skillContent string

const skillName = "debug-with-user"

const help = `
Usage: debug-with-user <command> [OPTIONS]

Commands:
  ask                      show a human checkpoint dialog (preset options + Customize)
  confirm                  alias for ask with --yes / --no options
  proceed                  alias for ask with --proceed / --cancel options
  skill show               print the embedded SKILL.md content
  skill install [OPTIONS]  install the skill to agent skill directories

Ask options:
  --title <text>           dialog title
  --message <text>         dialog message (supports \n)
  --option <label>         preset button (repeatable)
  --affirm <label>         which preset option counts as affirmed/pass
  --cancel <label>         cancel button label (default: Cancel)
  -h, --help               show help
`

type answerJSON struct {
	Answer   string `json:"answer"`
	Via      string `json:"via"`
	Affirmed *bool  `json:"affirmed,omitempty"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		os.Exit(0)
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Print(help)
	case "skill":
		if err := handleSkillCommand(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "debug-with-user: %v\n", err)
			os.Exit(1)
		}
	case "ask", "confirm", "proceed":
		code := runAskCommand(args[0], args[1:])
		os.Exit(code)
	default:
		fmt.Fprintf(os.Stderr, "debug-with-user: unknown command %q\n", args[0])
		fmt.Print(help)
		os.Exit(2)
	}
}

func handleSkillCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected skill show or skill install")
	}
	switch args[0] {
	case "show":
		if len(args) > 1 {
			return fmt.Errorf("unexpected arguments after show")
		}
		fmt.Print(skillContent)
		return nil
	case "install":
		return install.HandleInstall(install.InstallOptions{
			SkillDirName: skillName,
			SkillContent: skillContent,
			Usage:        "debug-with-user skill install",
		}, args[1:])
	default:
		return fmt.Errorf("unknown skill sub-command: %s, expected skill show or skill install", args[0])
	}
}

func runAskCommand(command string, args []string) int {
	req, err := parseAskArgs(command, args)
	if err != nil {
		if err == flags.ErrHelp {
			return 0
		}
		fmt.Fprintf(os.Stderr, "debug-with-user: %v\n", err)
		return 2
	}

	result, err := dialog.Ask(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "debug-with-user: %v\n", err)
		return 2
	}
	if result.Via == "dismissed" {
		fmt.Fprintln(os.Stderr, "debug-with-user: dialog dismissed")
		return 1
	}

	payload := answerJSON{
		Answer: result.Answer,
		Via:    result.Via,
	}
	if result.Via == "button" {
		payload.Affirmed = result.Affirmed
	}

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "debug-with-user: %v\n", err)
		return 2
	}
	fmt.Println(string(data))
	return 0
}

func parseAskArgs(command string, args []string) (dialog.AskRequest, error) {
	var title, message, affirm, cancel *string
	var options []string
	var yes, no, proceed *string

	remaining, err := flags.
		String("--title", &title).
		String("--message", &message).
		StringSlice("--option", &options).
		String("--affirm", &affirm).
		String("--cancel", &cancel).
		String("--yes", &yes).
		String("--no", &no).
		String("--proceed", &proceed).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return dialog.AskRequest{}, err
	}
	if len(remaining) > 0 {
		return dialog.AskRequest{}, fmt.Errorf("unexpected argument: %s", remaining[0])
	}

	req := dialog.AskRequest{}
	if title != nil {
		req.Title = unescapeNewlines(*title)
	}
	if message != nil {
		req.Message = unescapeNewlines(*message)
	}

	switch command {
	case "confirm":
		if yes != nil {
			req.Options = append(req.Options, unescapeNewlines(*yes))
		}
		if no != nil {
			req.Options = append(req.Options, unescapeNewlines(*no))
		}
		if affirm == nil && yes != nil {
			affirm = yes
		}
	case "proceed":
		if proceed != nil {
			req.Options = append(req.Options, unescapeNewlines(*proceed))
		}
		if cancel != nil {
			req.Options = append(req.Options, unescapeNewlines(*cancel))
		}
		if affirm == nil && proceed != nil {
			affirm = proceed
		}
	default:
		for _, opt := range options {
			req.Options = append(req.Options, unescapeNewlines(opt))
		}
	}

	if affirm != nil {
		req.AffirmOption = unescapeNewlines(*affirm)
	}
	if cancel != nil {
		req.CancelOption = unescapeNewlines(*cancel)
	}
	return req, nil
}

func unescapeNewlines(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n")
}