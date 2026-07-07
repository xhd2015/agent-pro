package main

import (
	"fmt"
	"strings"

	brainstorm_run "github.com/xhd2015/agent-pro/agents/brainstorm/run"
	debugwithuser_run "github.com/xhd2015/agent-pro/agents/debug-with-user/run"
	explore_run "github.com/xhd2015/agent-pro/agents/explore/run"
	followup_run "github.com/xhd2015/agent-pro/agents/followup/run"
	gitresolveconflicts_run "github.com/xhd2015/agent-pro/agents/git-resolve-conflicts/run"
	investigate_run "github.com/xhd2015/agent-pro/agents/investigate/run"
	intentroute_run "github.com/xhd2015/agent-pro/agents/intent-route/run"
	reproduce_run "github.com/xhd2015/agent-pro/agents/reproduce/run"
	verifywithprototype_run "github.com/xhd2015/agent-pro/agents/verify-with-prototype/run"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/skills/install"
)

type skillInfo struct {
	Name        string
	Description string
	Content     string
}

var knownSkills = map[string]skillInfo{
	"brainstorm": {
		Name:        "brainstorm",
		Description: extractDescription(brainstorm_run.SkillFile),
		Content:     brainstorm_run.SkillFile,
	},
	"debug-with-user": {
		Name:        "debug-with-user",
		Description: extractDescription(debugwithuser_run.SkillFile),
		Content:     debugwithuser_run.SkillFile,
	},
	"explore": {
		Name:        "explore",
		Description: extractDescription(explore_run.SkillFile),
		Content:     explore_run.SkillFile,
	},
	"followup": {
		Name:        "followup",
		Description: extractDescription(followup_run.SkillFile),
		Content:     followup_run.SkillFile,
	},
	"git-resolve-conflicts": {
		Name:        "git-resolve-conflicts",
		Description: extractDescription(gitresolveconflicts_run.SkillFile),
		Content:     gitresolveconflicts_run.SkillFile,
	},
	"investigate": {
		Name:        "investigate",
		Description: extractDescription(investigate_run.SkillFile),
		Content:     investigate_run.SkillFile,
	},
	"intent-route": {
		Name:        "intent-route",
		Description: extractDescription(intentroute_run.SkillFile),
		Content:     intentroute_run.SkillFile,
	},
	"reproduce": {
		Name:        "reproduce",
		Description: extractDescription(reproduce_run.SkillFile),
		Content:     reproduce_run.SkillFile,
	},
	"verify-with-prototype": {
		Name:        "verify-with-prototype",
		Description: extractDescription(verifywithprototype_run.SkillFile),
		Content:     verifywithprototype_run.SkillFile,
	},
}

func extractDescription(skillMD string) string {
	// Parse YAML frontmatter to extract description
	// The frontmatter is between --- markers
	rest := skillMD
	if !strings.HasPrefix(rest, "---\n") {
		return ""
	}
	rest = rest[4:] // skip "---\n"
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return ""
	}
	frontmatter := rest[:endIdx]

	lines := strings.Split(frontmatter, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			desc = strings.TrimSpace(desc)
			// Handle YAML folded block scalar (>-)
			// The description continues on subsequent indented lines
			if desc == ">-" || desc == ">" {
				// Collect continuation lines (indented)
				var parts []string
				for j := i + 1; j < len(lines); j++ {
					if lines[j] == "" || (lines[j][0] != ' ' && lines[j][0] != '\t') {
						break
					}
					parts = append(parts, strings.TrimSpace(lines[j]))
				}
				if len(parts) > 0 {
					return parts[0]
				}
				return "(multiline description)"
			}
			return desc
		}
	}
	return ""
}

func knownSkillNames() []string {
	return []string{"brainstorm", "debug-with-user", "explore", "followup", "git-resolve-conflicts", "intent-route", "investigate", "reproduce", "verify-with-prototype"}
}

const knownSkillNamesText = "brainstorm, debug-with-user, explore, followup, git-resolve-conflicts, intent-route, investigate, reproduce, verify-with-prototype"

const skillHelp = `
Usage: agent-pro skill <command> [ARGS]

Commands:
  --list, -l         list all available skill names
  <name> show         print the SKILL.md content of a skill
  <name> install      install a skill to a skill directory

Skill names: brainstorm, debug-with-user, explore, followup, git-resolve-conflicts, intent-route, investigate, reproduce, verify-with-prototype

Run agent-pro skill <name> <command> --help for command-specific options.
`

const skillsHelp = `
Usage: agent-pro skills [<command>] [ARGS]

Commands (without arguments, lists all available skill names):
  update              update already-installed skills
  <name> show         print the SKILL.md content of a skill
  <name> install      install a skill to a skill directory

Skill names: brainstorm, debug-with-user, explore, followup, git-resolve-conflicts, intent-route, investigate, reproduce, verify-with-prototype

Run agent-pro skills <name> <command> --help for command-specific options.
`

func handleSkill(args []string) error {
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(skillHelp, "\n"))
		return nil
	}

	// If first arg looks like a flag (starts with -), parse flags
	if strings.HasPrefix(args[0], "-") {
		var listFlag bool
		remaining, err := flags.Bool("-l,--list", &listFlag).
			Help("-h,--help", skillHelp).
			Parse(args)
		if err != nil {
			return err
		}
		if listFlag {
			return listSkills()
		}
		if len(remaining) > 0 {
			return fmt.Errorf("unexpected argument: %s", remaining[0])
		}
		return nil
	}

	name := args[0]
	sk, ok := knownSkills[name]
	if !ok {
		return fmt.Errorf("unknown skill: %s (available: %s)", name, knownSkillNamesText)
	}

	if len(args) == 1 {
		return fmt.Errorf("expected show or install after skill name %q", name)
	}

	switch args[1] {
	case "show":
		return handleSkillShow(sk, args[2:])
	case "install":
		return handleSkillInstall(sk, args[2:])
	case "-h", "--help":
		printSkillSubHelp(name)
		return nil
	default:
		return fmt.Errorf("unknown skill command: %s (expected show or install)", args[1])
	}
}

func handleSkills(args []string) error {
	if len(args) == 0 {
		return listSkills()
	}
	if args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(skillsHelp, "\n"))
		return nil
	}
	if args[0] == "update" {
		return handleSkillsUpdate(args[1:])
	}

	// If first arg is a skill name, delegate to handleSkill
	name := args[0]
	if _, ok := knownSkills[name]; ok {
		return handleSkill(args)
	}

	return fmt.Errorf("unknown skills command: %s (expected update or one of: %s)", args[0], knownSkillNamesText)
}

func listSkills() error {
	fmt.Println("Available skills:")
	for _, name := range knownSkillNames() {
		sk := knownSkills[name]
		if sk.Description != "" {
			fmt.Printf("  %-15s %s\n", name, sk.Description)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}

func printSkillSubHelp(name string) {
	fmt.Printf("Usage: agent-pro skill %s <command> [ARGS]\n\n", name)
	fmt.Printf("Commands:\n")
	fmt.Printf("  show                print the SKILL.md content\n")
	fmt.Printf("  install [OPTIONS]   install the skill\n\n")
	fmt.Printf("Install Options:\n")
	fmt.Printf("  --cursor            install to .cursor/skills/%s/\n", name)
	fmt.Printf("  --codex             install to .codex/skills/%s/\n", name)
	fmt.Printf("  --opencode          install to .opencode/skills/%s/\n", name)
	fmt.Printf("  --general-agents    install to .agents/skills/%s/\n", name)
	fmt.Printf("  --global            install to ~/.<dir>/... instead of ./\n")
	fmt.Printf("  --no-override       do not automatically overwrite existing\n")
	fmt.Printf("  --dry-run           show what would be created\n")
}

func handleSkillShow(sk skillInfo, args []string) error {
	_, err := flags.
		Help("-h,--help", fmt.Sprintf(`
Usage: agent-pro skill %s show

Print the SKILL.md content of the %s skill.

Options:
  -h,--help     show help
`, sk.Name, sk.Name)).
		Parse(args)
	if err != nil {
		return err
	}
	fmt.Print(sk.Content)
	return nil
}

func handleSkillInstall(sk skillInfo, args []string) error {
	return install.HandleInstall(install.InstallOptions{
		SkillDirName: sk.Name,
		SkillContent: sk.Content,
		Usage:        fmt.Sprintf("agent-pro skill %s install", sk.Name),
	}, args)
}

func handleSkillsUpdate(args []string) error {
	skills := make([]install.UpdateSkill, 0, len(knownSkills))
	for _, name := range knownSkillNames() {
		sk := knownSkills[name]
		skills = append(skills, install.UpdateSkill{
			InstallOptions: install.InstallOptions{
				SkillDirName: sk.Name,
				SkillContent: sk.Content,
				Usage:        "agent-pro skills update",
			},
			Name: name,
		})
	}
	return install.HandleUpdateMany(skills, args)
}
