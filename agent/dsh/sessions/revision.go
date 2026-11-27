package sessions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Revision is an opaque token that changes whenever a session's artifact changes.
//
// A reader that caches a projection of the log needs one cheap answer to "is the
// file I read still the same file", without knowing the log's path, encoding or
// byte offsets. The token folds those three into one string: a new format
// generation, an append, or a rewrite that kept the size all change it, and an
// idle session keeps it.
//
// ok is false when the session has no readable artifact.
func (s *Store) Revision(id string) (string, bool) {
	ref, err := s.StatLog(id)
	if err != nil {
		return "", false
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\n%d\n%d\n%d",
		ref.Path, ref.FormatVersion, ref.Size, ref.ModTimeUnixNano)))
	return hex.EncodeToString(sum[:16]), true
}

// Title answers the session's latest title from a bounded read of the log's tail.
//
// A title is one `session/title` row near the end of the artifact and the latest
// one wins, so folding the whole log to find it is what makes a board scan over a
// 200MB session cost minutes. SyncTail reads a suffix on a cold cursor and only
// the appended frames afterwards, which answers the same question for a fixed
// cost per poll.
//
// ok is false when the tail does not carry a title (an old title, or no
// artifact), and the caller then falls back to the folded Title consumer.
func (s *Store) Title(id string) (string, bool) {
	tail, err := s.SyncTail(id, nil)
	if err != nil || tail == nil {
		return "", false
	}
	title := strings.TrimSpace(tail.Summary.Title)
	return title, title != ""
}

// Size answers how many bytes the session's artifact holds, without decoding it.
//
// It is the one fact a reader needs to say "this log is large" without knowing
// where the log is or how it is encoded.
func (s *Store) Size(id string) (int64, bool) {
	ref, err := s.StatLog(id)
	if err != nil {
		return 0, false
	}
	return ref.Size, true
}
