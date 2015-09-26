package physical

type Internal_Index int

const (
	Loopback_Internal_Index = Internal_Index(1)
	External_Internal_Index = Internal_Index(2)
)

const TAP_NAME = "tap0"
const RX_QUEUE_SIZE = 5000
const MAX_FRAME_SZ = 1526
