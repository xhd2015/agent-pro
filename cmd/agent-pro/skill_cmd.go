package main

import (
	"fmt"
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

var verifyOnBehalfOfUserExtraFiles = func() []skillcmd.InstallFile {
	files, err := verifyonbehalfofuser_run.InstallFiles()
	if err != nil {
		panic("verify-on-behalf-of-user InstallFiles: " + err.Error())
	}
	return files
}()

// agentProSkills lists the skills agent-pro hosts, in help and --list order.
// A skill is one SKILL.md, optionally extended by a path/TOPIC.md tree
// (verify-on-behalf-of-user) and extra install files (its scripts/templates).
func agentProSkills() []skillcmd.RegisteredSkill {
	return []skillcmd.RegisteredSkill{
		{Name: "brainstorm", Content: brainstorm_run.SkillFile},
		{Name: "consolidate-code", Content: consolidatecode_run.SkillFile},
		{Name: "debug-with-user", Content: debugwithuser_run.SkillFile},
		{Name: "establish-a-loop", Content: establishaloop_run.SkillFile},
		{Name: "explore", Content: explore_run.SkillFile},
		{Name: "followup", Content: followup_run.SkillFile},
		{Name: "git-resolve-conflicts", Content: gitresolveconflicts_run.SkillFile},
		{Name: "intent-route", Content: intentroute_run.SkillFile},
		{Name: "investigate", Content: investigate_run.SkillFile},
		{Name: "reproduce", Content: reproduce_run.SkillFile},
		{Name: "run-the-loop", Content: runtheloop_run.SkillFile},
		{Name: "sound-fix", Content: soundfix_run.SkillFile},
		{Name: "split-phases", Content: splitphases_run.SkillFile},
		{Name: "summarize-a-skill", Content: summarizeaskill_run.SkillFile},
		{
			Name:       "verify-on-behalf-of-user",
			Content:    verifyonbehalfofuser_run.SkillFile,
			TreeFS:     verifyonbehalfofuser_run.SkillTree(),
			ExtraFiles: verifyOnBehalfOfUserExtraFiles,
		},
		{Name: "verify-with-prototype", Content: verifywithprototype_run.SkillFile},
	}
}

func agentProSkillRegistry() *skillcmd.Registry {
	return &skillcmd.Registry{
		Skills:      agentProSkills(),
		Usage:       "agent-pro skill --install",
		UpdateUsage: "agent-pro skills update",
		Help:        strings.TrimPrefix(skillHelp, "\n"),
		SkillsHelp:  strings.TrimPrefix(skillsHelp, "\n"),
	}
}

const skillHelp = `
Usage: agent-pro skill --list
       agent-pro skill --show <name> [<topic>]
       agent-pro skill <name> [--show] [<topic>]
       agent-pro skill --show <name>/<topic>
       agent-pro skill --version [<name>]
       agent-pro skill --install <name> [OPTIONS] [<dir>]
       agent-pro skill <name> --install [OPTIONS] [<dir>]

Actions (exactly one):
  --list, -l         list all available skill names
  --show             print the SKILL.md content of a skill
  --install          install a skill to a skill directory
  --version          print the skill's metadata.version
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

// handleSkill implements the multi-skill host actions as flags
// (--show / --install / --list), both arg orders.
func handleSkill(args []string) error {
	reg := agentProSkillRegistry()
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(skillHelp, "\n"))
		return nil
	}
	return reg.HandleSkill(args)
}

func handleSkills(args []string) error {
	reg := agentProSkillRegistry()
	if len(args) == 0 {
		return reg.HandleSkill([]string{"--list"})
	}
	return reg.HandleSkills(args)
}
