package block

import (
	"bytes"
	"fmt"

	"github.com/quic-go/quic-go/internal/fec"
	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/wire"
	"github.com/quic-go/quic-go/logging"
)

// TODO: E should be > 2

type BlockFrameworkSender struct {
	fecScheme                       BlockFECScheme
	redundancyController            RedundancyController
	fecFramesParser                 FECFramesParser
	currentBlock                    *FECBlock
	e                               protocol.ByteCount
	protectedPacketsSinceLastRepair []int
	nSourceSymbolsSinceLastRepair   int

	BlocksToSend []*FECBlock

	tracer *logging.ConnectionTracer
}

func NewBlockFrameworkSender(fecScheme BlockFECScheme, redundancyController RedundancyController, repairFrameParser FECFramesParser, E protocol.ByteCount, tracer *logging.ConnectionTracer) (*BlockFrameworkSender, error) {
	if E >= protocol.MAX_FEC_SYMBOL_SIZE {
		return nil, fmt.Errorf("framework sender symbol size too big: %d > %d", E, protocol.MAX_FEC_SYMBOL_SIZE)
	}
	if E <= 2 {
		return nil, fmt.Errorf("framework sender symbol size too small: %d <= 2", E)
	}
	return &BlockFrameworkSender{
		fecScheme:            fecScheme,
		redundancyController: redundancyController,
		fecFramesParser:      repairFrameParser,
		currentBlock:         NewFECBlock(0),
		e:                    E,
		tracer:               tracer,
	}, nil
}

var _ fec.FrameworkSender = &BlockFrameworkSender{}

func (f *BlockFrameworkSender) E() protocol.ByteCount {
	return protocol.ByteCount(f.e)
}

func (f *BlockFrameworkSender) GetNextFPID() protocol.SourceFECPayloadID {
	return BlockSourceID{
		BlockNumber: f.currentBlock.BlockNumber,
		BlockOffset: BlockOffset(len(f.currentBlock.sourceSymbolsOffsets)),
	}.ToFPID()
}

func (f *BlockFrameworkSender) protectSourceSymbol(symbol *BlockSourceSymbol) (retval protocol.SourceFECPayloadID) {
	buf := bytes.NewBuffer(nil)
	f.currentBlock.AddSourceSymbol(symbol).EncodeBlockSourceID(buf)
	copy(retval[:], buf.Bytes())
	return retval
}

// returns the ID of the first symbol in the payload
func (f *BlockFrameworkSender) ProtectPayload(pn protocol.PacketNumber, payload fec.PreProcessedPayload) (retval protocol.SourceFECPayloadID, err error) {
	fmt.Println("PROTECTPAYLOAD HAS BEEN CALLED")
	if payload == nil || len(payload.Bytes()) == 0 {
		return retval, fmt.Errorf("asked to protect an empty payload")
	}
	symbols, err := PayloadToSourceSymbols(payload.Bytes(), f.e, true)
	if err != nil {
		return retval, err
	}
	for i, symbol := range symbols {
		if i == 0 {
			retval = f.protectSourceSymbol(symbol)
		} else {
			f.protectSourceSymbol(symbol)
		}
	}

	f.protectedPacketsSinceLastRepair = append(f.protectedPacketsSinceLastRepair, len(symbols))
	f.nSourceSymbolsSinceLastRepair += len(symbols)
	// if f.redundancyController.ShouldSend(len(f.protectedPacketsSinceLastRepair)) {
	if true {
		err := f.GenerateRepairSymbols(f.currentBlock, f.redundancyController.GetNumberOfRepairSymbols(f.nSourceSymbolsSinceLastRepair))
		if err != nil {
			return retval, err
		}
		f.sendCurrentBlock()
	} else {
		fmt.Println("not sending yet")
	}
	return retval, nil
}

func (f *BlockFrameworkSender) sendCurrentBlock() {
	fmt.Println("!!! sending current block !!!")
	// f.tracer.Debug("fec", "block sent")
	f.currentBlock.TotalNumberOfSourceSymbols = uint64(len(f.currentBlock.SourceSymbols))
	f.currentBlock.TotalNumberOfRepairSymbols = uint64(len(f.currentBlock.RepairSymbols))
	f.BlocksToSend = append(f.BlocksToSend, f.currentBlock)

	f.currentBlock = NewFECBlock(f.currentBlock.BlockNumber + 1)
	f.protectedPacketsSinceLastRepair = f.protectedPacketsSinceLastRepair[:0]
	f.nSourceSymbolsSinceLastRepair = 0
}

func (f *BlockFrameworkSender) FlushUnprotectedSymbols() error {
	err := f.GenerateRepairSymbols(f.currentBlock, f.redundancyController.GetNumberOfRepairSymbols(f.nSourceSymbolsSinceLastRepair))
	if err != nil {
		return err
	}
	f.sendCurrentBlock()
	return nil
}

func (f *BlockFrameworkSender) GenerateRepairSymbols(block *FECBlock, numberOfSymbols uint) error {
	symbols, err := f.fecScheme.GetRepairSymbols(block, numberOfSymbols)
	if err != nil {
		return err
	}
	block.SetRepairSymbols(symbols)
	return nil
}

func (f *BlockFrameworkSender) GetRepairFrame(maxSize protocol.ByteCount) (*wire.RepairFrame, error) {
	if len(f.BlocksToSend) == 0 {
		return nil, nil
	}
	// find first block with at least one repair symbol
	for len(f.BlocksToSend[0].RepairSymbols) == 0 {
		// skip this block
		f.BlocksToSend = f.BlocksToSend[1:]
	}
	if len(f.BlocksToSend) == 0 {
		return nil, nil
	}

	block := f.BlocksToSend[0]
	rf, consumed, err := f.fecFramesParser.getRepairFrame(block, maxSize)
	if err != nil {
		return nil, err
	}
	block.RepairSymbols = block.RepairSymbols[consumed:]
	// if the fecBlock has been emptied by the parser, remove it
	if len(f.BlocksToSend) > 0 && len(f.BlocksToSend[0].RepairSymbols) == 0 {
		f.BlocksToSend = f.BlocksToSend[1:]
	}
	return rf, nil
}

func (f *BlockFrameworkSender) HandleRecoveredFrame(rf *wire.RecoveredFrame) ([]protocol.PacketNumber, error) {
	return f.fecFramesParser.getRecoveredFramePacketNumbers(rf)
}
