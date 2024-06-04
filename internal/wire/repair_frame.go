package wire

import (
	// "bytes"
	"github.com/quic-go/quic-go/internal/protocol"
)

type RepairFrame struct {
	Metadata      []byte
	RepairSymbols []byte
}

func (f *RepairFrame) Append(b []byte, version protocol.Version) ([]byte, error) {
	b = append(b, protocol.REPAIR_FRAME_TYPE)

	b = append(b, f.Metadata...)

	b = append(b, f.RepairSymbols...)

	return b, nil
}

// Length of a written frame
func (f *RepairFrame) Length(version protocol.Version) protocol.ByteCount {
	return protocol.ByteCount(1 + len(f.Metadata) + len(f.RepairSymbols))
}
