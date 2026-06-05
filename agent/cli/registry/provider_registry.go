package registry

import (
	"fmt"
	"strings"
)

type AgentRunnerInfo struct {
	ID   AgentRunnerID `json:"id"`
	Name string        `json:"name"`
}

type AgentRunner struct {
	ID    AgentRunnerID
	Name  string
	Agent Agent
}

type AgentRunnerRegistry struct {
	defaultID string
	ordered   []AgentRunner
	byID      map[AgentRunnerID]AgentRunner
}

func NewAgentRunnerRegistry(defaultID string, runners []AgentRunner) (*AgentRunnerRegistry, error) {
	reg := &AgentRunnerRegistry{
		byID: make(map[AgentRunnerID]AgentRunner, len(runners)),
	}
	for _, runner := range runners {
		id := AgentRunnerID(strings.TrimSpace(string(runner.ID)))
		if id == "" {
			return nil, fmt.Errorf("agent runner id is required")
		}
		if runner.Agent == nil {
			return nil, fmt.Errorf("agent runner %q is missing implementation", id)
		}
		runner.ID = id
		if strings.TrimSpace(runner.Name) == "" {
			runner.Name = string(id)
		}
		if _, exists := reg.byID[id]; exists {
			return nil, fmt.Errorf("duplicate agent runner id: %s", id)
		}
		reg.byID[id] = runner
		reg.ordered = append(reg.ordered, runner)
	}
	if len(reg.ordered) == 0 {
		return nil, fmt.Errorf("at least one agent runner is required")
	}
	if strings.TrimSpace(defaultID) == "" {
		reg.defaultID = string(reg.ordered[0].ID)
		return reg, nil
	}
	if _, ok := reg.byID[AgentRunnerID(defaultID)]; !ok {
		return nil, fmt.Errorf("default agent runner not found: %s", defaultID)
	}
	reg.defaultID = defaultID
	return reg, nil
}

func (r *AgentRunnerRegistry) DefaultID() string {
	if r == nil {
		return ""
	}
	return r.defaultID
}

func (r *AgentRunnerRegistry) Resolve(id string) (AgentRunner, error) {
	if r == nil {
		return AgentRunner{}, fmt.Errorf("agent runner registry is not configured")
	}
	runnerID := AgentRunnerID(strings.TrimSpace(id))
	if runnerID == "" {
		runnerID = AgentRunnerID(r.defaultID)
	}
	runner, ok := r.byID[runnerID]
	if !ok {
		return AgentRunner{}, fmt.Errorf("agent runner not found: %s", runnerID)
	}
	return runner, nil
}

func (r *AgentRunnerRegistry) List() []AgentRunnerInfo {
	if r == nil {
		return nil
	}
	out := make([]AgentRunnerInfo, 0, len(r.ordered))
	for _, runner := range r.ordered {
		out = append(out, AgentRunnerInfo{
			ID:   runner.ID,
			Name: runner.Name,
		})
	}
	return out
}
