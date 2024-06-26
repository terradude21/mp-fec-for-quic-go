package block

import (
	"fmt"

	"github.com/quic-go/quic-go/internal/fec"
	"github.com/quic-go/quic-go/internal/protocol"
)

type SynchronizationByte byte

func (b SynchronizationByte) IsPacketNumberPresent() bool {
	return b&0x04 == 0x04
}

func (b SynchronizationByte) IsStartOfPacket() bool {
	return b&0x02 == 0x02
}

func (b SynchronizationByte) IsEndOfPacket() bool {
	return b&0x01 == 0x01
}

func (b SynchronizationByte) SetPacketNumberPresent() SynchronizationByte {
	return b | 0x04
}

func (b SynchronizationByte) SetStartOfPacket() SynchronizationByte {
	return b | 0x02
}

func (b SynchronizationByte) SetEndOfPacket() SynchronizationByte {
	return b | 0x01
}

type BlockSourceSymbol struct {
	fec.SourceSymbol
	SynchronizationByte SynchronizationByte
	PacketChunk         []byte
}

func ParseBlockSourceSymbol(symbol []byte) *BlockSourceSymbol {
	return &BlockSourceSymbol{
		SourceSymbol: fec.SourceSymbol{
			Data: symbol,
		},
		SynchronizationByte: SynchronizationByte(symbol[0]),
		PacketChunk:         symbol[1:],
	}
}

func SourceSymbolToBlockSourceSymbol(symbol *fec.SourceSymbol) *BlockSourceSymbol {
	return &BlockSourceSymbol{
		SourceSymbol:        *symbol,
		SynchronizationByte: SynchronizationByte(symbol.Data[0]),
		PacketChunk:         symbol.Data[1:],
	}
}

func PayloadToSourceSymbols(payload []byte, E protocol.ByteCount, packetNumberPresent bool) ([]*BlockSourceSymbol, error) {
	packetChunkSize := int(E - 1)
	var retVal []*BlockSourceSymbol
	if len(payload)%packetChunkSize != 0 {
		return nil, fmt.Errorf("payload size (%d) is not aligned with E (%d)", len(payload), E)
	}
	totalLen := len(payload)
	nChunks := totalLen / packetChunkSize
	if nChunks*packetChunkSize < totalLen {
		nChunks++
	}
	for i := 0; i < nChunks; i++ {
		data := make([]byte, E)
		if i == 0 {
			if packetNumberPresent {
				data[0] = byte(SynchronizationByte(data[0]).SetPacketNumberPresent())
			}
			data[0] = byte(SynchronizationByte(data[0]).SetStartOfPacket())
		}
		if i == nChunks-1 {
			data[0] = byte(SynchronizationByte(data[0]).SetEndOfPacket())
		}
		if 0 < i && i < nChunks-1 {
			data[0] = 0
		}
		copy(data[1:], payload[:packetChunkSize])
		payload = payload[packetChunkSize:]
		retVal = append(retVal, ParseBlockSourceSymbol(data))
	}
	return retVal, nil
}
