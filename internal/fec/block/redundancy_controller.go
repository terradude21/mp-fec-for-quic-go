package block

import (
	"math"

	"github.com/quic-go/quic-go/internal/fec"
	"github.com/quic-go/quic-go/internal/protocol"
)

const DEFAULT_K = 5
const DEFAULT_N = 6

// The redundancy control will adapt the number of FEC Repair Symbols and
// the size of the FEC Block to the current conditions.

type RedundancyController interface {
	fec.RedundancyController
	// returns true if these symbols should be sent and protected with repair symbols
	// each element of the slide contains a slice of the symbols sent in a single packet
	ShouldSend(nPacketsSinceLastRepair int) bool
}

type constantRedundancyController struct {
	nRepairSymbols uint
	nSourceSymbols uint
	windowStepSize uint
}

var _ fec.RedundancyController = &constantRedundancyController{}

func NewConstantRedundancyController(nSourceSymbols uint, nRepairSymbols uint, windowStepSize uint) fec.RedundancyController {
	return &constantRedundancyController{
		nSourceSymbols: nSourceSymbols,
		nRepairSymbols: nRepairSymbols,
	}
}

func NewDefaultRedundancyController() fec.RedundancyController {
	return &constantRedundancyController{
		nSourceSymbols: DEFAULT_K,
		nRepairSymbols: DEFAULT_N - DEFAULT_K,
	}
}

func (*constantRedundancyController) OnSourceSymbolLost(pn protocol.PacketNumber) {}

func (*constantRedundancyController) OnSourceSymbolReceived(pn protocol.PacketNumber) {}

func (*constantRedundancyController) ShouldSend(nPacketsSinceLastRepair int) bool {
	// protect when K packets have been sent
	return nPacketsSinceLastRepair >= DEFAULT_K
}

func (c *constantRedundancyController) GetNumberOfRepairSymbols(nSymbolsSinceLastRepair int) uint {
	return uint(math.Round((float64(DEFAULT_N-DEFAULT_K)/float64(DEFAULT_N))*float64(nSymbolsSinceLastRepair))) + 1
}
