package types

type BroadcastStrategy string

const (
	BroadcastImmediate BroadcastStrategy = "IMMEDIATE"
	BroadcastPrivate   BroadcastStrategy = "PRIVATE_MEMPOOL"
	BroadcastManual    BroadcastStrategy = "MANUAL"
)
