package sessions

import "fmt"

const zstdMagic uint32 = 0xFD2FB528

type zstdFrame struct {
	start, end int
}

// scanZstdFrames locates complete concatenated Zstandard frames. A torn
// final frame returns tornStart and only the complete prefix.
func scanZstdFrames(buf []byte) (frames []zstdFrame, tornStart int, torn bool, err error) {
	offset := 0
	for offset < len(buf) {
		start := offset
		if len(buf)-offset < 4 {
			return frames, start, true, nil
		}
		magic := uint32(buf[offset]) | uint32(buf[offset+1])<<8 | uint32(buf[offset+2])<<16 | uint32(buf[offset+3])<<24
		if magic != zstdMagic {
			return nil, 0, false, fmt.Errorf("corrupt Zstandard session log: invalid frame magic at byte %d", offset)
		}
		offset += 4
		if offset == len(buf) {
			return frames, start, true, nil
		}
		descriptor := buf[offset]
		offset++
		if descriptor&0x18 != 0 {
			return nil, 0, false, fmt.Errorf("corrupt Zstandard session log: reserved frame-header bit at byte %d", offset-1)
		}
		contentSizeFlag := descriptor >> 6
		singleSegment := descriptor&0x20 != 0
		checksum := descriptor&0x04 != 0
		dictionaryFlag := descriptor & 0x03
		dictionaryBytes := 0
		if dictionaryFlag == 3 {
			dictionaryBytes = 4
		} else {
			dictionaryBytes = int(dictionaryFlag)
		}
		contentSizeBytes := 0
		if contentSizeFlag == 0 {
			if singleSegment {
				contentSizeBytes = 1
			}
		} else {
			contentSizeBytes = 1 << contentSizeFlag
		}
		remainingHeader := dictionaryBytes + contentSizeBytes
		if !singleSegment {
			remainingHeader++
		}
		if len(buf)-offset < remainingHeader {
			return frames, start, true, nil
		}
		offset += remainingHeader
		for {
			if len(buf)-offset < 3 {
				return frames, start, true, nil
			}
			blockHeader := uint32(buf[offset]) | uint32(buf[offset+1])<<8 | uint32(buf[offset+2])<<16
			offset += 3
			lastBlock := blockHeader&1 != 0
			blockType := (blockHeader >> 1) & 0x03
			blockSize := int(blockHeader >> 3)
			if blockType == 0x03 {
				return nil, 0, false, fmt.Errorf("corrupt Zstandard session log: reserved block type at byte %d", offset-3)
			}
			payloadBytes := blockSize
			if blockType == 0x01 {
				payloadBytes = 1
			}
			if len(buf)-offset < payloadBytes {
				return frames, start, true, nil
			}
			offset += payloadBytes
			if lastBlock {
				break
			}
		}
		if checksum {
			if len(buf)-offset < 4 {
				return frames, start, true, nil
			}
			offset += 4
		}
		frames = append(frames, zstdFrame{start: start, end: offset})
	}
	return frames, 0, false, nil
}
