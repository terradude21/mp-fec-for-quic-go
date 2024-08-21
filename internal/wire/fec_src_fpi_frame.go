package wire

import (
	// "bytes"

	"github.com/quic-go/quic-go/internal/protocol"
)

// TODO: this frame should be FECScheme (or FECFramework)-specific

// A FECSrcFPIFrame identifies a source symbol
type FECSrcFPIFrame struct {
	protocol.SourceFECPayloadID
}

func parseFECSrcFPIFrame(b []byte) (*FECSrcFPIFrame, int, error) {
	// if _, err := r.ReadByte(); err != nil {
	// 	return nil, 0, err
	// }

	frame := &FECSrcFPIFrame{}
	l := copy(frame.SourceFECPayloadID[:], b)
	return frame, l, nil
}

func (f *FECSrcFPIFrame) Append(b []byte, version protocol.Version) ([]byte, error) {
	b = append(b, protocol.FEC_SRC_FPI_FRAME_TYPE)
	b = append(b, f.SourceFECPayloadID[:]...)
	return b, nil
}

// Length of a written frame
func (f *FECSrcFPIFrame) Length(version protocol.Version) protocol.ByteCount {
	return protocol.ByteCount(1 + len(f.SourceFECPayloadID))
}
