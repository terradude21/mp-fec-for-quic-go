package fec

import (
	"bytes"

	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/utils"
	"github.com/quic-go/quic-go/internal/wire"
)

type FrameworkSender interface {
	// see coding-for-quic: e is the size of a source/repair symbol
	E() protocol.ByteCount
	ProtectPayload(number protocol.PacketNumber, payload PreProcessedPayload) (retval protocol.SourceFECPayloadID, err error)
	GetNextFPID() protocol.SourceFECPayloadID
	FlushUnprotectedSymbols() error
	GetRepairFrame(maxSize protocol.ByteCount) (*wire.RepairFrame, error)
	HandleRecoveredFrame(frame *wire.RecoveredFrame) ([]protocol.PacketNumber, error)
}

type FrameworkReceiver interface {
	E() protocol.ByteCount
	ReceivePayload(number protocol.PacketNumber, payload PreProcessedPayload, sourceID protocol.SourceFECPayloadID) error
	HandleRepairFrame(frame *wire.RepairFrame) error
	GetRecoveredPacket() *RecoveredPacket
	GetRecoveredFrame(maxLen protocol.ByteCount) (*wire.RecoveredFrame, error)
}

type PreProcessedPayload interface {
	Bytes() []byte
}

type preProcessedPayload struct {
	data []byte
}

func (p *preProcessedPayload) Bytes() []byte {
	return p.data
}

func PreparePayloadForEncoding(pn protocol.PacketNumber, framesToMaybeProtect []wire.Frame, sender FrameworkSender, version protocol.Version) (PreProcessedPayload, error) {
	data, err := preprocessPayload(pn, framesToMaybeProtect, protocol.ByteCount(sender.E()), version)
	if err != nil {
		return nil, err
	}
	return &preProcessedPayload{
		data: data,
	}, nil
}

func ReceivePayloadForDecoding(pn protocol.PacketNumber, framesToMaybeProtect []wire.Frame, receiver FrameworkReceiver, version protocol.Version) (PreProcessedPayload, error) {
	data, err := preprocessPayload(pn, framesToMaybeProtect, protocol.ByteCount(receiver.E()), version)
	if err != nil {
		return nil, err
	}
	return &preProcessedPayload{
		data: data,
	}, nil
}

func shouldProtect(f wire.Frame) bool {
	switch f.(type) {
	case *wire.AckFrame, *wire.RepairFrame, *wire.FECSrcFPIFrame, *wire.CryptoFrame:
		return false
	}
	return true
}

func writeProtectedFrames(frames []wire.Frame, version protocol.Version) ([]byte, error) {
	b := []byte{}
	var err error
	for _, f := range frames {
		if shouldProtect(f) {
			if b, err = f.Append(b, version); err != nil {
				return nil, err
			}
		}
	}
	return b, nil
}

func preprocessPayload(pn protocol.PacketNumber, framesToMaybeProtect []wire.Frame, E protocol.ByteCount, version protocol.Version) ([]byte, error) {
	payloadToProtect, err := writeProtectedFrames(framesToMaybeProtect, version)
	if err != nil {
		return nil, err
	}
	if len(payloadToProtect) == 0 {
		return nil, nil
	}
	packetChunkSize := E - 1
	lenWithoutPadding := utils.VarIntLen(uint64(pn)) + protocol.ByteCount(len(payloadToProtect))
	totalLen := lenWithoutPadding
	if (totalLen)%packetChunkSize != 0 {
		// align the length with packetChunkSize
		totalLen = (totalLen/packetChunkSize + 1) * packetChunkSize
	}
	// start writing after the padding
	b := bytes.NewBuffer(nil)
	utils.WriteVarInt(b, uint64(pn))
	// We leave this space full of zeroes: these are PADDING frames
	b.Write(bytes.Repeat([]byte{0}, int(totalLen-lenWithoutPadding)))
	//b.Next(int(totalLen - lenWithoutPadding))
	b.Write(payloadToProtect)
	// now, the payload is aligned with packetChunkSize(). It contains padding frames, then the full packet number as a VarInt, then the
	// payload to protect
	return b.Bytes(), nil
}

type RecoveredPacket struct {
	Number  protocol.PacketNumber
	Payload []byte
}
