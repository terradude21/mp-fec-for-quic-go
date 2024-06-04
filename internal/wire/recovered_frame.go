package wire

import (
	// "bytes"
	"github.com/quic-go/quic-go/internal/protocol"
)

// the RECOVERED frame format is defined by the underlying FEC Framework/Scheme
type RecoveredFrame struct {
	Data []byte
}

func (f *RecoveredFrame) Append(b []byte, version protocol.Version) ([]byte, error) {
	b = append(b, protocol.RECOVERED_FRAME_TYPE)

	b = append(b, f.Data...)

	return b, nil
}

// Length of a written frame
func (f *RecoveredFrame) Length(version protocol.Version) protocol.ByteCount {
	return protocol.ByteCount(1 + len(f.Data))
}
