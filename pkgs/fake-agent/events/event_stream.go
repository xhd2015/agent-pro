package events

import (
	"sync"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/fake-agent/probe"
)

// EventStream yields AgentEvents one at a time using the same logic and RNG
// sequence as GenerateEvents, without running the full probe loop upfront.
type EventStream struct {
	mu                sync.Mutex
	g                 *Generator
	prompt            string
	topic             string
	probeList         []probe.Suggestion
	round             int
	thinkDone         bool
	probesInitialized bool
	messageDone       bool
	// skipProbeLoop, when set, yields think then message with no intervening
	// tool probes. Used by llm-mock HTTP fallback so clients (e.g. grok) are
	// not blocked on bash/grep execution against a large repo cwd.
	skipProbeLoop bool
}

// NewEventStream creates a lazy generator for the given seed and prompt.
func NewEventStream(seed int64, prompt string) *EventStream {
	return &EventStream{
		g:      NewGenerator(seed),
		prompt: prompt,
	}
}

// NewMockEventStream creates a fast lazy generator for llm-mock: think on the
// first Next(), message on the second, with no synchronous probe execution.
func NewMockEventStream(seed int64, prompt string) *EventStream {
	return &EventStream{
		g:             NewGenerator(seed),
		prompt:        prompt,
		skipProbeLoop: true,
	}
}

// Next returns the next generated event, or false when the session is exhausted.
func (es *EventStream) Next() (types.AgentEvent, bool) {
	es.mu.Lock()
	defer es.mu.Unlock()
	return es.nextLocked()
}

// PeekNext returns the next event without advancing stream state.
func (es *EventStream) PeekNext() (types.AgentEvent, bool) {
	es.mu.Lock()
	defer es.mu.Unlock()

	savedThinkDone := es.thinkDone
	savedMessageDone := es.messageDone
	savedTopic := es.topic
	savedProbeList := es.probeList
	savedRound := es.round
	savedProbesInitialized := es.probesInitialized

	evt, ok := es.nextLocked()

	es.thinkDone = savedThinkDone
	es.messageDone = savedMessageDone
	es.topic = savedTopic
	es.probeList = savedProbeList
	es.round = savedRound
	es.probesInitialized = savedProbesInitialized

	return evt, ok
}

// IsMockFast reports whether this stream uses the llm-mock fast path (think then message).
func (es *EventStream) IsMockFast() bool {
	es.mu.Lock()
	defer es.mu.Unlock()
	return es.skipProbeLoop
}

func (es *EventStream) nextLocked() (types.AgentEvent, bool) {
	if !es.thinkDone {
		es.thinkDone = true
		es.topic = extractTopic(es.prompt)
		thinkID := es.g.nextID()
		return es.g.generateThink(thinkID, es.topic), true
	}

	if es.skipProbeLoop {
		if !es.messageDone {
			es.messageDone = true
			msgID := es.g.nextID()
			return es.g.generateMessage(msgID, es.topic), true
		}
		return types.AgentEvent{}, false
	}

	if !es.probesInitialized {
		es.probeList = probe.Scan(es.prompt)
		es.probeList = probe.Merge(es.probeList, probe.DefaultSuggestions())
		es.probesInitialized = true
	}

	if es.round < genMaxRounds && !es.messageDone {
		if es.g.chance(genResponseChance) {
			es.round = genMaxRounds
		} else {
			suggestion := es.g.pickSuggestion(es.probeList)
			toolEvent, result := es.g.execProbe(suggestion)
			es.round++
			if result != "" {
				newProbes := probe.Scan(result)
				es.probeList = probe.Merge(es.probeList, newProbes)
			}
			return toolEvent, true
		}
	}

	if !es.messageDone {
		es.messageDone = true
		msgID := es.g.nextID()
		return es.g.generateMessage(msgID, es.topic), true
	}

	return types.AgentEvent{}, false
}