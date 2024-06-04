package fec

import "github.com/quic-go/quic-go/internal/protocol"

type RepairSymbol struct {
	FECSchemeID protocol.FECSchemeID
	data        []byte // May contain metadata
}
