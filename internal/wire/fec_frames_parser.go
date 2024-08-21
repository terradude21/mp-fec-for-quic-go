package wire

type FECFramesParser interface {
	ParseRecoveredFrame(b []byte) (*RecoveredFrame, int, error)
	ParseRepairFrame(b []byte) (*RepairFrame, int, error)
}
