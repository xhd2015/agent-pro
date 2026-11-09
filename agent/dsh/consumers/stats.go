package consumers

// StatsID is the consumer that accumulates the Messages bar's numbers.
const StatsID = "stats"

// StatsState is the whole-session tally the Messages bar shows. It is not the
// paged window: compaction and "load earlier" never change these totals, because
// they are folded from the durable log rather than from the rows on screen.
//
// The counters are deliberately not UsageState. Usage answers "what did this
// session cost", where input, output, and cache read are separate and cache
// writes are not part of the sum; the footer answers "how big did this session
// get", where billed input includes cache writes and the last prompt is what
// fills the window. One state cannot be both without one number lying.
type StatsState struct {
	// Turns counts user messages, which is the footer's "turns": a turn is a
	// thing the user asked for.
	Turns int64 `json:"turns,omitempty"`
	// Tokens is billed input plus output over every assistant message.
	Tokens int64 `json:"tokens,omitempty"`
	// CacheRead is summed cache reads, the footnote's hit-rate numerator.
	CacheRead int64 `json:"cacheRead,omitempty"`
	// BilledInput is summed billed input, its denominator.
	BilledInput int64 `json:"billedInput,omitempty"`
	// LastPromptTokens is the billed input of the last assistant message: what
	// the next request starts from.
	LastPromptTokens int64 `json:"lastPromptTokens,omitempty"`
	// MessageTokens is the billed input of the last message that had one.
	MessageTokens int64 `json:"messageTokens,omitempty"`
}

// SessionStats folds the rows the footer counts.
//
// Payload keys are the wire's, not the fold's: events carry inputTokens,
// outputTokens, cacheReadTokens, and cacheWriteTokens. A fold that reads the
// shorter names silently counts zero cache reads, which is how a 95% hit rate
// turns into "cache 0%".
func SessionStats() Consumer {
	return &Typed[StatsState]{
		Name:  StatsID,
		Ver:   1,
		Watch: []string{"user/message", "assistant/message"},
		Apply: func(state *StatsState, ev Event, _ Point) error {
			switch ev.Type {
			case "user/message":
				state.Turns++
				return nil
			case "assistant/message":
				usage := messageUsage(ev.Data)
				if usage == nil {
					return nil
				}
				billed := billedInput(usage)
				state.CacheRead += usage.CacheRead
				state.BilledInput += billed
				state.Tokens += billed + usage.Output
				if billed > 0 {
					state.LastPromptTokens = billed
					state.MessageTokens = billed
				}
				return nil
			}
			return nil
		},
	}
}

// usageCounts is one assistant message's token report.
type usageCounts struct {
	Input      int64
	Output     int64
	CacheRead  int64
	CacheWrite int64
	Total      int64
}

// billedInput is what a request paid for, whichever way the row reports it: the
// parts when the input count is a total, the total when the parts disagree.
func billedInput(usage *usageCounts) int64 {
	uncached := usage.Input
	if usage.CacheRead > 0 && usage.Input >= usage.CacheRead {
		uncached = usage.Input - usage.CacheRead
	}
	billed := uncached + usage.CacheRead + usage.CacheWrite
	if usage.Total > billed+usage.Output {
		billed = usage.Total - usage.Output
	}
	return billed
}

// messageUsage reads an assistant message's usage object, under either the
// flat or the nested spelling a log can use.
func messageUsage(raw []byte) *usageCounts {
	fields := decodeObject(raw)
	if fields == nil {
		return nil
	}
	usage := objectField(fields, "usage")
	if usage == nil {
		usage = objectField(objectField(fields, "message"), "usage")
	}
	if usage == nil {
		return nil
	}
	counts := &usageCounts{
		Input:      integer(usage, "inputTokens"),
		Output:     integer(usage, "outputTokens"),
		CacheRead:  integer(usage, "cacheReadTokens"),
		CacheWrite: integer(usage, "cacheWriteTokens"),
		Total:      integer(usage, "totalTokens"),
	}
	if counts.Input == 0 && counts.Output == 0 && counts.CacheRead == 0 && counts.Total == 0 {
		return nil
	}
	return counts
}

// Stats returns the folded tally.
func (s Snapshot) Stats() StatsState {
	var state StatsState
	_ = s.Value(StatsID, &state)
	return state
}

// Empty reports that the session never reported anything the footer shows.
func (s StatsState) Empty() bool {
	return s.Turns == 0 && s.Tokens == 0 && s.LastPromptTokens == 0
}
