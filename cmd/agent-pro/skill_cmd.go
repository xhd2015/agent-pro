package main

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	brainstorm_run "github.com/xhd2015/agent-pro/agents/brainstorm/run"
	consolidatecode_run "github.com/xhd2015/agent-pro/agents/consolidate-code/run"
	debugwithuser_run "github.com/xhd2015/agent-pro/agents/debug-with-user/run"
	establishaloop_run "github.com/xhd2015/agent-pro/agents/establish-a-loop/run"
	explore_run "github.com/xhd2015/agent-pro/agents/explore/run"
	followup_run "github.com/xhd2015/agent-pro/agents/followup/run"
	gitresolveconflicts_run "github.com/xhd2015/agent-pro/agents/git-resolve-conflicts/run"
	investigate_run "github.com/xhd2015/agent-pro/agents/investigate/run"
	intentroute_run "github.com/xhd2015/agent-pro/agents/intent-route/run"
	runtheloop_run "github.com/xhd2015/agent-pro/agents/run-the-loop/run"
	reproduce_run "github.com/xhd2015/agent-pro/agents/reproduce/run"
	soundfix_run "github.com/xhd2015/agent-pro/agents/sound-fix/run"
	splitphases_run "github.com/xhd2015/agent-pro/agents/split-phases/run"
	summarizeaskill_run "github.com/xhd2015/agent-pro/agents/summarize-a-skill/run"
	verifyonbehalfofuser_run "github.com/xhd2015/agent-pro/agents/verify-on-behalf-of-user/run"
	verifywithprototype_run "github.com/xhd2015/agent-pro/agents/verify-with-prototype/run"
	"github.com/xhd2015/skills/skillcmd"
)

type skillInfo struct {
	Name        string
	Description string
	Content     string
	TreeFS      fs.FS // optional nested path/TOPIC.md tree
	ExtraFiles  []skillcmd.InstallFile
}

var verifyOnBehalfOfUserExtraFiles = func() []skillcmd.InstallFile {
	files, err := verifyonbehalfofuser_run.InstallFiles()
	if err != nil {
		panic("verify-on-behalf-of-user InstallFiles: " + err.Error())
	}
	return files
}()

var knownSkills = map[string]skillInfo{
	"brainstorm": {
		Name:        "brainstorm",
		Description: extractDescription(brainstorm_run.SkillFile),
		Content:     brainstorm_run.SkillFile,
	},
	"consolidate-code": {
		Name:        "consolidate-code",
		Description: extractDescription(consolidatecode_run.SkillFile),
		Content:     consolidatecode_run.SkillFile,
	},
	"debug-with-user": {
		Name:        "debug-with-user",
		Description: extractDescription(debugwithuser_run.SkillFile),
		Content:     debugwithuser_run.SkillFile,
	},
	"establish-a-loop": {
		Name:        "establish-a-loop",
		Description: extractDescription(establishaloop_run.SkillFile),
		Content:     establishaloop_run.SkillFile,
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
	"run-the-loop": {
		Name:        "run-the-loop",
		Description: extractDescription(runtheloop_run.SkillFile),
		Content:     runtheloop_run.SkillFile,
	},
	"reproduce": {
		Name:        "reproduce",
		Description: extractDescription(reproduce_run.SkillFile),
		Content:     reproduce_run.SkillFile,
	},
	"sound-fix": {
		Name:        "sound-fix",
		Description: extractDescription(soundfix_run.SkillFile),
		Content:     soundfix_run.SkillFile,
	},
	"split-phases": {
		Name:        "split-phases",
		Description: extractDescription(splitphases_run.SkillFile),
		Content:     splitphases_run.SkillFile,
	},
	"summarize-a-skill": {
		Name:        "summarize-a-skill",
		Description: extractDescription(summarizeaskill_run.SkillFile),
		Content:     summarizeaskill_run.SkillFile,
	},
	"verify-on-behalf-of-user": {
		Name:        "verify-on-behalf-of-user",
		Description: extractDescription(verifyonbehalfofuser_run.SkillFile),
		Content:     verifyonbehalfofuser_run.SkillFile,
		TreeFS:      verifyonbehalfofuser_run.SkillTree(),
		ExtraFiles:  verifyOnBehalfOfUserExtraFiles,
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
	return []string{"brainstorm", "consolidate-code", "debug-with-user", "establish-a-loop", "explore", "followup", "git-resolve-conflicts", "intent-route", "investigate", "reproduce", "run-the-loop", "sound-fix", "split-phases", "summarize-a-skill", "verify-on-behalf-of-user", "verify-with-prototype"}
}

const knownSkillNamesText = "brainstorm, consolidate-code, debug-with-user, establish-a-loop, explore, followup, git-resolve-conflicts, intent-route, investigate, reproduce, run-the-loop, sound-fix, split-phases, summarize-a-skill, verify-on-behalf-of-user, verify-with-prototype"

const skillHelp = `
Usage: agent-pro skill --list
       agent-pro skill --show <name> [<topic>]
       agent-pro skill <name> [--show] [<topic>]
       agent-pro skill --show <name>/<topic>
       agent-pro skill --install <name> [OPTIONS] [<dir>]
       agent-pro skill <name> --install [OPTIONS] [<dir>]

Actions (exactly one):
  --list, -l         list all available skill names
  --show             print the SKILL.md content of a skill
  --install          install a skill to a skill directory
  --header           with --show, print YAML frontmatter only

Skill names: brainstorm, consolidate-code, debug-with-user, establish-a-loop, explore, followup, git-resolve-conflicts, intent-route, investigate, reproduce, run-the-loop, sound-fix, split-phases, summarize-a-skill, verify-on-behalf-of-user, verify-with-prototype

Run agent-pro skill --install <name> --help for install options (--global, --cursor, …).
`

const skillsHelp = `
Usage: agent-pro skills
       agent-pro skills update [OPTIONS] [<dir>]
       agent-pro skills --show <name>
       agent-pro skills <name> --show
       agent-pro skills --install <name> [OPTIONS] [<dir>]
       agent-pro skills <name> --install [OPTIONS] [<dir>]

With no arguments, list all available skills.
skills update refreshes already-installed skills only.
skills also accepts the same --show / --install flag actions as skill.

Skill names: brainstorm, consolidate-code, debug-with-user, establish-a-loop, explore, followup, git-resolve-conflicts, intent-route, investigate, reproduce, run-the-loop, sound-fix, split-phases, summarize-a-skill, verify-on-behalf-of-user, verify-with-prototype

Run agent-pro skills update --help for update options.
Run agent-pro skill --help for skill actions.
`

// handleSkill implements Shape 2 multi-skill host actions as flags
// (--show / --install / --list), both arg orders. See go-best-practice skill-cli.
func handleSkill(args []string) error {
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(skillHelp, "\n"))
		return nil
	}

	parsed, err := skillcmd.ParseSkillArgs(args)
	if err != nil {
		return err
	}
	switch parsed.Action {
	case skillcmd.ActionHelp:
		fmt.Print(strings.TrimPrefix(skillHelp, "\n"))
		return nil
	case skillcmd.ActionList:
		return listSkills()
	case skillcmd.ActionShow:
		return handleSkillShow(parsed.Header, parsed.Rest)
	case skillcmd.ActionInstall:
		return handleSkillInstall(parsed.Rest)
	default:
		return fmt.Errorf("unknown action: %s", parsed.Action)
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
	// skills <name> --show / skills --install <name> … same as skill
	return handleSkill(args)
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

func handleSkillShow(header bool, rest []string) error {
	name, topicPath, extra, err := parseSkillShowArgs(rest)
	if err != nil {
		return err
	}
	if len(extra) > 0 {
		return fmt.Errorf("unexpected arguments: %v", extra)
	}
	sk, ok := knownSkills[name]
	if !ok {
		return fmt.Errorf("unknown skill: %s (available: %s)", name, knownSkillNamesText)
	}
	content := sk.Content
	if topicPath != "" {
		if sk.TreeFS == nil {
			return fmt.Errorf("unknown topic path: %s for skill %s", topicPath, name)
		}
		content, err = loadSkillTopicFromTree(sk.TreeFS, topicPath)
		if err != nil {
			return err
		}
	}
	if header {
		out, err := skillcmd.FormatHeaderWithDelimiters(content)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	}
	fmt.Print(content)
	return nil
}

func parseSkillShowArgs(rest []string) (name, topicPath string, extra []string, err error) {
	if len(rest) == 0 {
		return "", "", nil, fmt.Errorf("expected skill name for --show (try --help)")
	}
	name = rest[0]
	extra = rest[1:]
	if i := strings.Index(name, "/"); i > 0 {
		topicPath = name[i+1:]
		name = name[:i]
		return name, topicPath, extra, nil
	}
	if len(extra) > 0 {
		topicPath = extra[0]
		extra = extra[1:]
	}
	return name, topicPath, extra, nil
}

func loadSkillTopicFromTree(treeFS fs.FS, topicPath string) (string, error) {
	topicPath = strings.Trim(topicPath, "/")
	if topicPath == "" {
		return "", fmt.Errorf("empty topic path")
	}
	for _, s := range strings.Split(topicPath, "/") {
		if s == "" || s == "." || s == ".." {
			return "", fmt.Errorf("invalid topic path segment: %q", s)
		}
	}
	data, err := fs.ReadFile(treeFS, path.Join(topicPath, "TOPIC.md"))
	if err != nil {
		return "", fmt.Errorf("unknown topic: %s", topicPath)
	}
	return string(data), nil
}

func handleSkillInstall(rest []string) error {
	if len(rest) == 0 {
		return fmt.Errorf("expected skill name for --install (try --help)")
	}
	name := rest[0]
	sk, ok := knownSkills[name]
	if !ok {
		return fmt.Errorf("unknown skill: %s (available: %s)", name, knownSkillNamesText)
	}
	return skillcmd.HandleInstall(skillcmd.InstallOptions{
		SkillDirName: sk.Name,
		SkillContent: sk.Content,
		ExtraFiles:   sk.ExtraFiles,
		Usage:        fmt.Sprintf("agent-pro skill --install %s", sk.Name),
	}, rest[1:])
}

func handleSkillsUpdate(args []string) error {
	skills := make([]skillcmd.UpdateSkill, 0, len(knownSkills))
	for _, name := range knownSkillNames() {
		sk := knownSkills[name]
		skills = append(skills, skillcmd.UpdateSkill{
			InstallOptions: skillcmd.InstallOptions{
				SkillDirName: sk.Name,
				SkillContent: sk.Content,
				ExtraFiles:   sk.ExtraFiles,
				Usage:        "agent-pro skills update",
			},
			Name: name,
		})
	}
	return skillcmd.HandleUpdateMany(skills, args)
}
