package sessions

import (
	"fmt"
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// zstdStreamDecoders reuses decoders for prefix reads. artifactID,
// readArtifact and recentSummary each decode the head of a different log, and
// a fresh decoder allocates its window for every call: decoding the 421
// headers of a cold store allocated 1332MB in 261ms, against 9MB in 46ms with
// this pool. It is a separate pool from zstdDecoders because a decoder that
// has started streaming cannot be handed to DecodeAll.
var zstdStreamDecoders = sync.Pool{New: func() any {
	dec, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil
	}
	return dec
}}

// zstdStream resets a pooled decoder onto r. The caller must call
// releaseZstdStream once the read is done. The returned decoder is owned by
// the caller until then.
func zstdStream(r io.Reader) (*zstd.Decoder, error) {
	raw := zstdStreamDecoders.Get()
	if raw == nil {
		return nil, fmt.Errorf("zstd: cannot create decoder")
	}
	dec := raw.(*zstd.Decoder)
	if err := dec.Reset(r); err != nil {
		zstdStreamDecoders.Put(dec)
		return nil, err
	}
	return dec, nil
}

// releaseZstdStream detaches the reader so the pooled decoder never keeps a
// closed file alive, then returns it with its window intact.
func releaseZstdStream(dec *zstd.Decoder) {
	if dec == nil {
		return
	}
	_ = dec.Reset(nil)
	zstdStreamDecoders.Put(dec)
}
