package fec_utils

import (
	"fmt"

	"github.com/quic-go/quic-go/internal/fec"
	"github.com/quic-go/quic-go/internal/fec/block"
	"github.com/quic-go/quic-go/internal/fec/block/fec_schemes"
	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/wire"
	"github.com/quic-go/quic-go/logging"
)

func CreateFrameworkSenderFromFECSchemeID(id protocol.FECSchemeID, controller fec.RedundancyController, symbolSize protocol.ByteCount, tracer *logging.ConnectionTracer) (fec.FrameworkSender, wire.FECFramesParser, error) {
	switch {
	case IsBlockFECScheme(id):
		fmt.Println("FEC IS GETTING SET UP 1")
		fecScheme, err := GetBlockFECScheme(id)
		if err != nil {
			return nil, nil, err
		}
		fmt.Println("FEC IS GETTING SET UP 2")
		if controller == nil {
			controller = block.NewDefaultRedundancyController()
		}
		if blockController, ok := controller.(block.RedundancyController); !ok {
			return nil, nil, fmt.Errorf("wrong redundancy controller: expected a BlockRedundancyController")
		} else {
			rfp := block.NewFECFramesParser(symbolSize)
			sender, err := block.NewBlockFrameworkSender(fecScheme, blockController, rfp, symbolSize, tracer)
			return sender, rfp, err
		}
	case id == protocol.FECDisabled:
		fmt.Println("FEC IS DISABLED")
		return nil, nil, nil
	default:
		return nil, nil, fmt.Errorf("invalid sender FECSchemeID: %d", id)
	}
}

func CreateFrameworkReceiverFromFECSchemeID(id protocol.FECSchemeID, symbolSize protocol.ByteCount) (fec.FrameworkReceiver, wire.FECFramesParser, error) {
	switch {
	case IsBlockFECScheme(id):
		fecScheme, err := GetBlockFECScheme(id)
		if err != nil {
			return nil, nil, err
		}
		rfp := block.NewFECFramesParser(symbolSize)
		receiver, err := block.NewBlockFrameworkReceiver(fecScheme, rfp, symbolSize)
		return receiver, rfp, err
	case id == protocol.FECDisabled:
		return nil, nil, nil
	default:
		return nil, nil, fmt.Errorf("invalid receiver FECSchemeID: %d", id)
	}
}

func IsBlockFECScheme(id protocol.FECSchemeID) bool {
	switch id {
	case protocol.XORFECScheme, protocol.ReedSolomonFECScheme:
		return true
	default:
		return false
	}
}

func GetBlockFECScheme(id protocol.FECSchemeID) (block.BlockFECScheme, error) {
	switch id {
	case protocol.XORFECScheme:
		return &fec_schemes.XORFECScheme{}, nil
	case protocol.ReedSolomonFECScheme:
		return fec_schemes.NewReedSolomonFECScheme()
	default:
		return nil, fmt.Errorf("invalid block FEC Scheme ID: %d", id)
	}
}
