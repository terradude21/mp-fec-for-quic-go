package wire

import (
	"bytes"

	"github.com/quic-go/quic-go/internal/protocol"
)

// TODO: this frame should be FECScheme (or FECFramework)-specific

// A FECSrcFPIFrame identifies a source symbol
type FECSrcFPIFrame struct {
	protocol.SourceFECPayloadID
}

func parseFECSrcFPIFrame(r *bytes.Reader) (*FECSrcFPIFrame, error) {
	if _, err := r.ReadByte(); err != nil {
		return nil, err
	}

	frame := &FECSrcFPIFrame{}
	if _, err := r.Read(frame.SourceFECPayloadID[:]); err != nil {
		return nil, err
	} else {
		return frame, nil
	}
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
