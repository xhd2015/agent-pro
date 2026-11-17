package sessions

import (
	"encoding/json"
	"fmt"
	"strings"
)

// compactCommand is the slash command that folds history by hand. An automatic
// compaction has no command: it logs its lifecycle and then appends the
// checkpoint message that replaces the range it summarized.
const compactCommand = "compact"

// compactionCheckpointKind tags the user/message a compaction appends in place
// of the shadowed range. The summary text itself is model-facing, so the
// checkpoint is the only thing in the transcript that says a fold happened.
const compactionCheckpointKind = "compact-checkpoint"

// compactionMark is the metering one compaction/summary event reported: how
// many surface rows it replaced and what they were priced at under the token
// meter's fixed estimator.
type compactionMark struct {
	items  int
	tokens int64
}

// compactionFold turns a compaction lifecycle into transcript rows. Events
// arrive in order and each one needs only what the previous ones left behind:
//
//   - command/run and command/done carry a manual /compact. The done event does
//     not repeat the command name, so the run's name is remembered by
//     commandId. They project to a pending row and a settled row that share a
//     ref, so a live reader can replace the first with the second instead of
//     showing a card that spins forever. The settled row keeps the done event's
//     seq: a reader polling afterSeq has already seen the run's seq, and a row
//     that kept it would never be delivered.
//   - compaction/summary holds the folded range's size, which the checkpoint
//     that follows does not repeat. It is remembered by compactionId.
//   - the checkpoint is an automatic compaction's only visible surface, so it
//     becomes the row itself. A checkpoint a manual command already owns
//     (sourceCommandId) is skipped: the command row speaks for it.
//
// compaction/start, compaction/end and compaction/prune project to nothing.
// They are log-only, and a prune's replacement is an ordinary tool/result that
// already has a row.
type compactionFold struct {
	names map[string]string
	marks map[string]compactionMark
}

// rows projects one event, or nothing when the event is not compaction's to
// show.
func (f *compactionFold) rows(event Event) []Message {
	switch event.Type {
	case "command/run":
		id, name := commandID(event), toolName(event)
		if name != compactCommand {
			return nil
		}
		if id != "" {
			if f.names == nil {
				f.names = map[string]string{}
			}
			f.names[id] = name
		}
		row := commandRunMessage(event)
		row.Ref = id
		return []Message{row}
	case "command/done":
		id := commandID(event)
		if f.names[id] != compactCommand {
			return nil
		}
		row := commandDoneMessage(event)
		// command/done does not repeat the name the run carried.
		row.Tool = compactCommand
		row.Ref = id
		return []Message{row}
	case "compaction/summary":
		f.remember(event)
		return nil
	case "user/message":
		checkpoint, ok := readCompactionCheckpoint(event)
		if !ok || checkpoint.commandID != "" {
			return nil
		}
		return []Message{{
			Seq:  event.Seq,
			Time: event.Time,
			Kind: TranscriptCommand,
			Text: f.line(checkpoint.compactionID),
			Sum:  "success",
			Ref:  checkpoint.compactionID,
		}}
	}
	return nil
}

// remember keeps one summary event's metering. A compaction whose summary fell
// outside the retained window leaves no mark, and the row says so by staying
// text-free instead of inventing counts.
func (f *compactionFold) remember(event Event) {
	var data struct {
		CompactionID       string  `json:"compactionId"`
		ShadowedSeqs       []int64 `json:"shadowedSeqs"`
		ShadowedTokenCount int64   `json:"shadowedTokenCount"`
	}
	if json.Unmarshal(event.Data, &data) != nil {
		return
	}
	id := strings.TrimSpace(data.CompactionID)
	if id == "" {
		return
	}
	if f.marks == nil {
		f.marks = map[string]compactionMark{}
	}
	f.marks[id] = compactionMark{items: len(data.ShadowedSeqs), tokens: data.ShadowedTokenCount}
}

// line names what one automatic checkpoint folded, in the numbers the log kept.
// Empty means the evidence is missing, not that the fold was empty.
func (f *compactionFold) line(compactionID string) string {
	mark, ok := f.marks[compactionID]
	if !ok || mark.items <= 0 {
		return ""
	}
	return fmt.Sprintf("%d history items folded · ~%d tokens", mark.items, mark.tokens)
}

// compactionCheckpoint is the replacement message a compaction appends, read
// down to the two identities a projection can use.
type compactionCheckpoint struct {
	compactionID string
	commandID    string
}

// readCompactionCheckpoint reports whether event is a compaction's replacement
// user/message.
func readCompactionCheckpoint(event Event) (compactionCheckpoint, bool) {
	if event.Type != "user/message" {
		return compactionCheckpoint{}, false
	}
	var data struct {
		Source struct {
			Kind            string `json:"kind"`
			CompactionID    string `json:"compactionId"`
			SourceCommandID string `json:"sourceCommandId"`
		} `json:"source"`
	}
	if json.Unmarshal(event.Data, &data) != nil {
		return compactionCheckpoint{}, false
	}
	if strings.TrimSpace(data.Source.Kind) != compactionCheckpointKind {
		return compactionCheckpoint{}, false
	}
	return compactionCheckpoint{
		compactionID: strings.TrimSpace(data.Source.CompactionID),
		commandID:    strings.TrimSpace(data.Source.SourceCommandID),
	}, true
}
